use rusqlite::Connection;
use regex::Regex; 
use serenity::{async_trait, client::Client, client::Context, client::EventHandler, model::channel::Message, model::gateway::Activity, model::gateway::Ready, prelude::TypeMapKey};
use serde::Deserialize;
use std::{collections::HashMap, fs::File, io::BufReader,  sync::Arc};
use tokio::sync::Mutex;

// TypeMapKeys /////////////////////////////////////////////////////////////////////////////////////////////
struct CommandMap;
impl TypeMapKey for CommandMap {
    type Value = Arc<HashMap<&'static str, Box<dyn Command>>>;
}

struct Dictionary;
impl TypeMapKey for Dictionary {
    type Value = Arc<HashMap<String, String>>;
}

struct ArgSplitter;
impl TypeMapKey for ArgSplitter {
    type Value = Arc<Regex>;
}

struct Prefix;
impl TypeMapKey for Prefix {
    type Value = Arc<String>;
}

struct AliasDatabase;
impl TypeMapKey for AliasDatabase {
    type Value = Arc<Mutex<Connection>>;
}


// Common functions /////////////////////////////////////////////////////////////////////////////////////
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



// Commands /////////////////////////////////////////////////////////////////////////////////////////////
#[async_trait]
pub trait Command: Send + Sync {
    async fn exec(&self, ctx: &Context, msg: &Message, args: &Vec<&str>);
}

struct CmdResist;
#[async_trait] impl Command for CmdResist {
    async fn exec(&self, ctx: &Context, msg: &Message, args: &Vec<&str>) {
        if args.len() != 5 || args[3] != "vs" {
            say(ctx, msg, wrap_code!(help_resist!())).await; 
            return;
        }
        
        let active = match args[2].parse::<i32>() {
            Ok(x) => x,
            Err(_) => {
                say(ctx, msg, wrap_code!(help_resist!())).await;
                return;
            }
        };

        let passive = match args[4].parse::<i32>() {
            Ok(x) => x,
            Err(_) => {
                say(ctx, msg, wrap_code!(help_resist!())).await;
                return;
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

    }
}


struct CmdVersion;
#[async_trait] impl Command for CmdVersion {
    async fn exec(&self, ctx: &Context, msg: &Message, _: &Vec<&str>) {
        say(ctx, msg, "I'm CocBot v1.0.0, written in Rust!!").await;
    }
}
struct CmdHelp;
#[async_trait] impl Command for CmdHelp {
    async fn exec(&self, ctx: &Context, msg: &Message, _: &Vec<&str>) {
        say(ctx, msg, wrap_code!(help!())).await;
    }
}


struct CmdFind; 
#[async_trait] impl Command for CmdFind {
    async fn exec(&self, ctx: &Context, msg: &Message, args: &Vec<&str>) {
        if args.len() <= 2 {
            say(ctx, msg, wrap_code!(help_find!())).await;
            return;
        }
        let key: String = args[2..].join(" ").to_lowercase();
        let mut aka_key: Option<String> = None;
        let data = ctx.data.read().await;
        {
            let alias_db = data.get::<AliasDatabase>()
                .expect("[CmdFind] AliasDatabase not set!")
                .lock()
                .await;
            let mut stmt = alias_db.prepare("SELECT value FROM alias WHERE key = (?)")
                .expect("[CmdFind] Problem preparing query");
        
            let mut rows = stmt.query(&[&key])
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

    }
}

struct CmdRemoveAlias;
#[async_trait] impl Command for CmdRemoveAlias {
    async fn exec(&self, ctx: &Context, msg: &Message, args: &Vec<&str>) {
        if args.len() <= 2 {
            say(ctx, msg, wrap_code!(help_alias!())).await;
            return;
        }
        let alias_name = args[2..].join(" ").to_lowercase();
        let rows_affected: usize;
        {
            let data = ctx.data.read().await;
            let alias_db = data.get::<AliasDatabase>()
                .expect("[CmdRemoveAlias] AliasDatabase not set!")
                .lock()
                .await;
            rows_affected = alias_db.execute("DELETE FROM alias WHERE key = (?)", &[&alias_name])
                .expect("[CmdRemoveAlias]  Cannot execute query!");
        }

        if rows_affected == 0 {
            say(ctx, msg, format!("Sorry, I can't find an alias named **{}**...", alias_name)).await;
            return;
        }
        
        say(ctx, msg, format!("Done! **{}** is not longer an alias! ^^b", alias_name)).await;

    }

}

struct CmdAddAlias;
#[async_trait] impl Command for CmdAddAlias {
    async fn exec(&self, ctx: &Context, msg: &Message, args: &Vec<&str>) {
        if args.len() <= 2 {
            say(ctx, msg, wrap_code!(help_alias!())).await;
            return;
        }
    
        let alias_name: &str;
        let target_name: &str;
        let str_to_parse = args[2..].join(" ").to_lowercase();
        {    
            let str_to_parse_arr = str_to_parse.split(" = ").collect::<Vec<&str>>();
            if str_to_parse_arr.len() != 2 {
                say(ctx, msg, wrap_code!(help_alias!())).await;
                return;
            }
            alias_name = str_to_parse_arr.get(0).expect("[CmdAddAlias] Problem getting alias_name");
            target_name = str_to_parse_arr.get(1).expect("[CmdAddAlias] Problem getting target_name");
        }   

        let rows_affected: usize;
        {
            let data = ctx.data.read().await;
            {
                let dictionary = data.get::<Dictionary>()
                    .expect("[CmdAddAlias] Dictionary not set!");
                if !dictionary.contains_key(target_name)  {
                    say(ctx, msg, "Target not found! Are you sure the target name is correct?").await;
                    return;
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
            return;
        }
        
        say(ctx, msg, format!("Alias added! **{}** is now also known as **{}**!", target_name, alias_name)).await;
    }
}

struct CmdGetAlias;
#[async_trait] impl Command for CmdGetAlias {
    async fn exec(&self, ctx: &Context, msg: &Message, args: &Vec<&str>) {
        if args.len() <= 2 {
            say(ctx, msg, wrap_code!(help_alias!())).await;
            return;
        }
        let mut found = false;
        let mut result: String = String::new();
        let alias_name: String = args[2..].join(" ");
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
    }
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
    // Set a handler for the `message` event - so that whenever a new message
    // is received - the closure (or function) passed will be called.
    //
    // Event handlers are dispatched through a threadpool, and so multiple
    // events can be dispatched simultaneously.
    async fn message(&self, ctx: Context, msg: Message) {
        let data = ctx.data.read().await;
        let prefix = data.get::<Prefix>().expect("Prefix not set!");
        // Handle prefix
        let is_prefixed = msg.content.starts_with(prefix.as_str());
        
        if is_prefixed {
            let arg_splitter = data.get::<ArgSplitter>().expect("ArgSplitter not set!");
            let command_map  = data.get::<CommandMap>().expect("CommandMap not set!");
            let mut arg_list: Vec<&str> = Vec::new();
            {
                for part in &mut arg_splitter.find_iter(msg.content.as_str()) {
                    arg_list.push(part.as_str());
                }
            }

            let cmd_str: &str;
            {
                let opt_cmd_str = arg_list.get(1);
                if opt_cmd_str.is_none() {
                    return;
                }
                cmd_str = opt_cmd_str.unwrap();
            }

            let command: &Box<dyn Command>;
            {
                let opt_command = command_map.get(cmd_str);
                if opt_command.is_none() {
                    return;
                }
                command = opt_command.unwrap();
            }
            command.exec(&ctx, &msg, &arg_list).await;
        }

    }

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

   

    // Create a new instance of the Client, logging in as a bot. This will
    // automatically prepend your bot token with "Bot ", which is a requirement
    // by Discord for bot users.
    let mut client = Client::new(&config.token)
        .event_handler(DiscordHandler)
        .await
        .expect("Error creating client");

    {
        let mut data = client.data.write().await;
        // Command map
        {
            let mut cmd_map: HashMap<&str, Box<dyn Command>> = HashMap::new();
            cmd_map.insert("version",  Box::new(CmdVersion{}));
            cmd_map.insert("get-alias", Box::new(CmdGetAlias{}));
            cmd_map.insert("add-alias", Box::new(CmdAddAlias{}));
            cmd_map.insert("remove-alias", Box::new(CmdRemoveAlias{}));
            cmd_map.insert("find", Box::new(CmdFind{}));
            cmd_map.insert("help", Box::new(CmdHelp{}));
            cmd_map.insert("resist", Box::new(CmdResist{}));
            data.insert::<CommandMap>(Arc::new(cmd_map));
        }

     
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

        data.insert::<ArgSplitter>(Arc::new(Regex::new(r#"(?i)(?:[^\s"]+\b|:|(")[^"]*("))+|[=!&|~]"#)
            .expect("Cannot initialize arg_splitter")));
    
        data.insert::<Prefix>(Arc::new(String::from(config.prefix)));

       
    }

    // Finally, start a single shard, and start listening to events.
    //
    // Shards will automatically attempt to reconnect, and will perform
    // exponential backoff until it reconnects.
    if let Err(why) = client.start().await {
        println!("Client error: {:?}", why);
    }
}
