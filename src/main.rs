use rusqlite::Connection;

use serenity::async_trait;
use serenity::client::Client; 
use serenity::client::Context; 
use serenity::client::EventHandler; 
use serenity::model::channel::Message; 
use serenity::model::gateway::Activity; 
use serenity::model::gateway::Ready; 
use serenity::prelude::TypeMapKey;
use serenity::framework::standard::macros::*;
use serenity::framework::standard::Args;
use serenity::framework::standard::StandardFramework;
use serenity::framework::standard::CommandResult;

use serde::Deserialize;

use std::collections::HashMap; 
use std::fs::File;
use std::io::BufReader;
use std::sync::Arc;
use tokio::sync::Mutex;

// TypeMapKeys ///////////////////////////////////////////////////////////////
struct Dictionary;
impl TypeMapKey for Dictionary {
    type Value = Arc<HashMap<String, String>>;
}

struct AliasDatabase;
impl TypeMapKey for AliasDatabase {
    type Value = Arc<Mutex<Connection>>;
}


// Common functions /////////////////////////////////////////////////////////////
macro_rules! help_find {
    () => ( "find: Use this command to find something\n\t> Usage: !coc find <something>\n" )
}

macro_rules! help_add_alias {
    () => ( "add-alias: Adds an alias to a 'find'\n\t> Usage: !coc add-alias <alias_name> = <target_name>\n\t(if success, you can then do '!necro, find <alias_name>')\n" )
}

macro_rules! help_remove_alias {
    () => ("remove-alias: Removes an alias\n\t> Usage: !coc remove-alias <alias_name>\n")
}

macro_rules! help_get_alias {
    () => ("get-alias: Displays an alias\n\t> Usage: !coc get-alias <alias_name>\n")
}

macro_rules! help_resist {
    () => ("resist: Check CoC resistance!\n\t> Usage: !coc resist <active> vs <passive>\n")
}


macro_rules! help_alias {
    () => (concat!(help_add_alias!(), help_remove_alias!(), help_get_alias!()));
}

macro_rules! help {
    () => (concat!(help_resist!(), help_find!(), help_add_alias!(), help_remove_alias!(), help_get_alias!()));
}

macro_rules! wrap_code {
    ($item:expr) => (concat!("```", $item, "```"))
}
async fn say(ctx: &Context, msg: &Message, display: impl std::fmt::Display)  {
    if let Err(why) = msg.channel_id.say(&ctx.http, display).await {
        println!("Error sending message: {:?}", why);
    }
}


fn args_to_string(mut args: Args) -> String {
    let mut ret = String::with_capacity(128);
    ret.push_str(args.single::<String>().unwrap().as_str());
    for arg in args.iter::<String>() {
        ret.push_str(format!(" {}", arg.unwrap()).as_str());
    }

    return ret;
}

// Commands /////////////////////////////////////////////////////////////
#[group]
#[commands(version, get_alias, add_alias, remove_alias, find, help, resist)]
struct General;


#[command]
async fn version(ctx: &Context, msg: &Message) -> CommandResult {
    say(ctx, msg, "I'm CocBot v2.0.0, written in Rust!!").await;
    return Ok(());
}

#[command]
async fn resist(ctx: &Context, msg: &Message, mut args: Args) -> CommandResult {
    if args.len() != 3 {
        say(ctx, msg, wrap_code!(help_resist!())).await; 
        return Ok(());
    }

    let active = match args.single::<i32>() {
        Ok(x) => x,
        Err(_) => {
            say(ctx, msg, wrap_code!(help_resist!())).await;
            return Ok(());
        }
    };

    match args.single::<String>() {
        Ok(x) => {
            if x != "vs" {            
                say(ctx, msg, wrap_code!(help_resist!())).await;
                return Ok(());
            }
        },
        Err(_) => {
            say(ctx, msg, wrap_code!(help_resist!())).await;
            return Ok(());
        }
    };

    let passive = match args.single::<i32>() {
        Ok(x) => x,
        Err(_) => {
            say(ctx, msg, wrap_code!(help_resist!())).await;
            return Ok(());
        }
    };
     
    let result = (active - passive) * 5 + 50;
    if result > 95 {
        say(ctx, msg, format!("Let's see...\nActive: **{}**\nPassive: **{}**\nThe result is an **Automatic Success**!!\\(^o^)/", active, passive)).await;
    }
    else if result < 5 {
        say(ctx, msg, format!("Let's see...\nActive: **{}**\nPassive: **{}**\nThe result is an **Automatic Failure**!! (´・ω・`)", active, passive)).await;
    }
    else {
        say(ctx, msg, format!("Let's see...\nActive: **{}**\nPassive: **{}**\nThe result is **{}**!! （｀・ω・´）", active, passive, result)).await;
    }

    return Ok(());
}

#[command]
async fn help(ctx: &Context, msg: &Message) -> CommandResult {
    say(ctx, msg, wrap_code!(help!())).await;
    return Ok(());
}

#[command]
async fn find(ctx: &Context, msg: &Message, args: Args) -> CommandResult {
    let key: String = args_to_string(args).to_lowercase();
    let mut aka_key: Option<String> = None;
    let data = ctx.data.read().await;
    {
        let alias_db = data.get::<AliasDatabase>()
            .expect("[CmdFind] AliasDatabase not set!")
            .lock()
            .await;
        let mut stmt = alias_db
            .prepare("SELECT value FROM alias WHERE key = (?)")
            .expect("[CmdFind] Problem preparing query");
    
        let mut rows = stmt
            .query(&[&key])
            .expect("[CmdFind] Problem executing query");
         
        if let Some(row) = rows.next().expect("[CmdFind] Problem getting row") {
            aka_key = Some(row.get(0).expect("[CmdFind] Problem getting value from row"));
        } 
    }
    
    let dictionary = data.get::<Dictionary>()
        .expect("[CmdFind] Dictionary not set!");

    match aka_key {
        Some(aka_key_v) => {
            match dictionary.get(aka_key_v.as_str()) {
                Some(value) => say(ctx, msg, format!("I found **{}** (aka **{}**)! ```{}```", key, aka_key_v, value)).await,
                None => say(ctx, msg, "Sorry...I can't find what you are looking for >_<").await
            };
        },
        None => {
            match dictionary.get(key.as_str()) {
                Some(value) => say(ctx, msg, format!("I found **{}**! ```{}```", key, value)).await,
                None => say(ctx, msg, "Sorry...I can't find what you are looking for >_<").await
            };
        }
    }
    return Ok(());
}

#[command]
#[aliases("remove-alias")]
async fn remove_alias(ctx: &Context, msg: &Message, args: Args) -> CommandResult {
    if args.len() == 0 {
        say(ctx, msg, wrap_code!(help_alias!())).await;
        return Ok(());
    }

    let alias_name: String = args_to_string(args).to_lowercase();
    let rows_affected: usize;
    {
        let data = ctx.data.read().await;
        let alias_db = data.get::<AliasDatabase>()
            .expect("[CmdRemoveAlias] AliasDatabase not set!")
            .lock()
            .await;
        rows_affected = alias_db
            .execute("DELETE FROM alias WHERE key = (?)", &[&alias_name])
            .expect("[CmdRemoveAlias]  Cannot execute query!");
    }

    if rows_affected == 0 {
        say(ctx, msg, format!("Sorry, I can't find an alias named **{}**...", alias_name)).await;
        return Ok(());
    }
    
    say(ctx, msg, format!("Done! **{}** is not longer an alias! ^^b", alias_name)).await;
    return Ok(());

}

#[command]
#[aliases("add-alias")]
async fn add_alias(ctx: &Context, msg: &Message, args: Args) -> CommandResult{
    if args.len() <= 2 {
        say(ctx, msg, wrap_code!(help_alias!())).await;
        return Ok(());
    }

    let alias_name: &str;
    let target_name: &str;
    let str_to_parse: String = args_to_string(args).to_lowercase();
    {    
        let str_to_parse_arr = str_to_parse.split(" = ").collect::<Vec<&str>>();
        if str_to_parse_arr.len() != 2 {
            say(ctx, msg, wrap_code!(help_alias!())).await;
            return Ok(());
        }
        alias_name = str_to_parse_arr.get(0)
            .expect("[CmdAddAlias] Problem getting alias_name");
        target_name = str_to_parse_arr.get(1)
            .expect("[CmdAddAlias] Problem getting target_name");
    }   

    let rows_affected: usize;
    {
        let data = ctx.data.read().await;
        {
            let dictionary = data.get::<Dictionary>()
                .expect("[CmdAddAlias] Dictionary not set!");
            if !dictionary.contains_key(target_name)  {
                say(ctx, msg, "Target not found! Are you sure the target name is correct?").await;
                return Ok(());
            }
        }

        let alias_db = data.get::<AliasDatabase>()
            .expect("[CmdAddAlias] AliasDatabase not set!")
            .lock()
            .await;
        rows_affected = alias_db.execute("INSERT OR IGNORE INTO alias VALUES (?, ?)", &[&alias_name, &target_name])
            .expect("[CmdAddAlias]  Cannot execute query!");
    }

    if rows_affected == 0 {
        say(ctx, msg, format!("Duplicate alias **{}** found! Please remove first with the *alias remove* command", alias_name)).await;
        return Ok(());
    }
    
    say(ctx, msg, format!("Alias added! **{}** is now also known as **{}**!", target_name, alias_name)).await;
    return Ok(());
}

#[command]
#[aliases("get-alias")]
async fn get_alias(ctx: &Context, msg: &Message, args: Args) -> CommandResult {
    if args.len() == 0 {
        say(ctx, msg, wrap_code!(help_alias!())).await;
        return Ok(());
    }
    let mut found = false;
    let mut result: String = String::new();    
    let alias_name: String = args_to_string(args).to_lowercase();
    {
        let data = ctx.data.read().await;
        let alias_db = data.get::<AliasDatabase>()
            .expect("[CmdGetAlias] AliasDatabase not set!")
            .lock()
            .await;
         
        let mut stmt = alias_db.prepare("SELECT value FROM alias WHERE key = (?)")
                .expect("[CmdGetAlias] Problem preparing query");
        
        let mut rows = stmt.query(&[&alias_name])
            .expect("[CmdGetAlias] Problem executing query");
      
        if let Some(row) = rows.next().expect("[CmdGetAlias] Problem getting row") {
            result = row.get(0).expect("[CmdGetAlias] Problem getting value from row");
            found = true;
        } 
    }
    
    match found {
        true => say(ctx, msg, format!("**{}** is also known as **{}**.", alias_name.as_str(), result.as_str())).await,
        false => say(ctx, msg, format!("Sorry, I can't find an alias named **{}**...", alias_name.as_str())).await,
    }
    return Ok(());
}


#[derive(Deserialize)]
struct Config {
    token: String,
    prefix: String,
    data_json_path: String,
    alias_db_path: String,
}

struct DiscordHandler; 
#[async_trait] impl EventHandler for DiscordHandler {
    // Set a handler to be called on the `ready` event. This is called when a
    // shard is booted, and a READY payload is sent by Discord. This payload
    // contains data like the current user's guild Ids, current user data,
    // private channels, and more.
    //
    // In this case, just print what the current user's username is.
    async fn ready(&self, ctx: Context, ready: Ready) {
        println!("{} is connected!", ready.user.name);
        ctx.set_activity(Activity::playing("type !coc help")).await;
    }
}



#[tokio::main]
async fn main() {

    let config: Config;
    {
        let file = File::open("config.json")
            .expect("Cannot open 'config.json'");
       
        let reader = BufReader::new(file);

        config = serde_json::from_reader(reader)
            .expect("Cannot parse 'config.json'");
    }


    let framework = StandardFramework::new()
        .configure(|c| c
                   .with_whitespace(true)
                   .prefix(config.prefix.as_str()))
        .group(&GENERAL_GROUP);

    // Create a new instance of the Client, logging in as a bot. This will
    // automatically prepend your bot token with "Bot ", which is a requirement
    // by Discord for bot users.
    let mut client = Client::builder(&config.token)
                        .event_handler(DiscordHandler)
                        .framework(framework)
                        .await
                        .unwrap();

    {
        let mut data = client.data.write().await;
     
        // alias database
        {
            let alias_conn = Connection::open(config.alias_db_path)
                .expect("Cannot open alias database");
            data.insert::<AliasDatabase>(Arc::new(Mutex::new(alias_conn)));
        }
        // data_json
        {
            let data_json: HashMap<String, String>;
            let file = File::open(config.data_json_path)
                .expect("Cannot open data_json_path");
            let reader = BufReader::new(file);
            data_json = serde_json::from_reader(reader)
                .expect("Cannot parse data_json_path");
            
            data.insert::<Dictionary>(Arc::new(data_json));
        }
       
    }

    // Finally, start a single shard, and start listening to events.
    //
    // Shards will automatically attempt to reconnect, and will perform
    // exponential backoff until it reconnects.
    if let Err(why) = client.start().await {
        println!("Client error: {:?}", why);
    }
}
