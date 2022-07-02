package main

import (
  "github.com/bwmarrin/discordgo"
  bolt "go.etcd.io/bbolt"

  "strings"
  "fmt"
  "regexp"
  "strconv"
)

var (
  global_command_arg_splitter  *regexp.Regexp
)

const DISCORD_MESSAGE_MAX_CHARACTERS int = 2000

func init_commands() {
  global_command_arg_splitter = regexp.MustCompile(`(?i)(?:[^\s"]+\b|:|(")[^"]*("))+|[=!&|~+\-\*\/\%]`)
  if global_command_arg_splitter  == nil {
    panik("arg_splitter failed to compile")
  }
}

func execute_commands(ses *discordgo.Session, msg *discordgo.MessageCreate) {
  args := global_command_arg_splitter.FindAllString(msg.Content, -1)
  if args != nil && len(args) >= 2 {
    cmd_str := args[1]
    args = args[2:] // get array from 2 to n

    // Use this instead of interfaces
    // Clearer, more concise, easier to debug
    switch cmd_str {
        case "help":
            ses.ChannelMessageSend(msg.ChannelID, MSG_HELP)
    case "version":
      ses.ChannelMessageSend(msg.ChannelID,   MSG_VERSION)
        case "add-alias":
            cmd_add_alias(ses, msg, args)
        case "get-alias":
            cmd_get_alias(ses, msg, args)
        case "remove-alias":
            cmd_remove_alias(ses, msg, args)
        case "find":
            cmd_find(ses, msg, args)
    case "resist":
            cmd_resist(ses, msg, args)
      }
    }
}

func wrap_code(str string) string {
  return "```" + str + "```"
}

func cmd_find(ses *discordgo.Session, msg *discordgo.MessageCreate, args []string) {
  if len(args) != 1 {
    reply := fmt.Sprintf(MSG_HELP_QUERY, MSG_HELP_FIND)
    ses.ChannelMessageSend(msg.ChannelID, reply)
    return
  }
  //defer recovery(discord, message)

  key := strings.ToLower(args[0])

  // Check if it is an alias. If so, get the actual key from alias.
  global_db.View(func(tx *bolt.Tx) error {
    bucket := tx.Bucket([]byte("alias")) 
    alias_value := bucket.Get([]byte(key))
    if alias_value != nil {
      // key is an alias 
      // Thus, find the entry from dictionary using alias value
      alias_value_str := string(alias_value)
      entry, exist := global_dictionary[alias_value_str]
      if !exist {
          ses.ChannelMessageSend(msg.ChannelID, MSG_FIND_FAIL)
          return nil
      }
      entry_str := string(entry)
      reply := fmt.Sprintf(MSG_FIND_PASS_WITH_ALIAS, key, alias_value_str, entry_str) 
      ses.ChannelMessageSend(msg.ChannelID, reply)
      return nil
    } 
     
    // key is an alias 
    // Thus, find the entry from dictionary using alias value
    entry, exist := global_dictionary[key]
    if !exist {
        ses.ChannelMessageSend(msg.ChannelID, MSG_FIND_FAIL)
        return nil
    }
    entry_str := string(entry)
    reply := fmt.Sprintf(MSG_FIND_PASS, key, entry_str) 
    ses.ChannelMessageSend(msg.ChannelID, reply)
    return nil
  })
}

func cmd_remove_alias(ses *discordgo.Session, msg *discordgo.MessageCreate, args []string) {
  if len(args) <= 0 {
    reply := fmt.Sprintf(MSG_HELP_QUERY, MSG_HELP_REMOVE_ALIAS)
    ses.ChannelMessageSend(msg.ChannelID, reply)
    return
  }

  alias_name := strings.Join(args[0:], " ")
  global_db.Update(func(tx *bolt.Tx) error {
    // Check if entry exists
    b := tx.Bucket([]byte("alias")) 
    v := b.Get([]byte(alias_name))
    if v == nil {
      // If it does not exist, we failed  
      reply := fmt.Sprintf(MSG_REMOVE_ALIAS_FAIL, alias_name)
      ses.ChannelMessageSend(msg.ChannelID, reply)  
    } else {
      // if it does, then there's a duplicate
      b.Delete([]byte(alias_name))
      reply := fmt.Sprintf(MSG_REMOVE_ALIAS_PASS, alias_name)
      ses.ChannelMessageSend(msg.ChannelID, reply) 
    }
    return nil
  })
}

func cmd_get_alias(ses *discordgo.Session, msg *discordgo.MessageCreate, args []string) {
  if len(args) <= 0 {
    reply := fmt.Sprintf(MSG_HELP_QUERY, MSG_HELP_GET_ALIAS)
    ses.ChannelMessageSend(msg.ChannelID, reply)
    return
  }

  alias_name := strings.Join(args[0:], " ")

  // Select value from alias where key = ?
  global_db.View(func(tx *bolt.Tx) error {
    // Check if entry exists
    b := tx.Bucket([]byte("alias")) 
    target_name := b.Get([]byte(alias_name))
    if target_name == nil {
      // Does not exist
      reply := fmt.Sprintf(MSG_GET_ALIAS_FAIL, alias_name)
      ses.ChannelMessageSend(msg.ChannelID, reply)
    } else {
      // If it does, then there's a duplicate
      reply := fmt.Sprintf(MSG_GET_ALIAS_PASS, target_name, alias_name)
      ses.ChannelMessageSend(msg.ChannelID, reply) 
    }
    return nil
  })
}

func cmd_add_alias(ses *discordgo.Session, msg *discordgo.MessageCreate, args []string) {
  if len(args) <= 0 {
    reply := fmt.Sprintf(MSG_HELP_QUERY, MSG_HELP_ADD_ALIAS)
    ses.ChannelMessageSend(msg.ChannelID, reply)
    return
  }
    
  // combine '<alias> = <target>' into one string
  str_to_parse := strings.Join(args[0:], " ")             

  // split to: '<alias>', '=', '<target>'
  str_to_parse_arr := strings.Split(str_to_parse, " = ")   
  if len(str_to_parse_arr) != 2 {                      
    // we expect 2 items in the array: '<alias>' and '<target>'
    reply := fmt.Sprintf(MSG_HELP_QUERY, MSG_HELP_ADD_ALIAS)
    ses.ChannelMessageSend(msg.ChannelID, reply)
    return
  }
  alias_name := strings.ToLower(str_to_parse_arr[0])
  target_name := strings.ToLower(str_to_parse_arr[1])
    
  if _, exist := global_dictionary[target_name]; !exist {
    ses.ChannelMessageSend(msg.ChannelID, MSG_ADD_ALIAS_TARGET_NOT_FOUND)
  } else {
    // TODO: We can optimize by using View to check 
    // if an alias exist first, then use Update if it doesn't.
    // Then again, it might be safer to just lock the whole thing.
    // Whatever #lazyprogramming
    global_db.Update(func(tx *bolt.Tx) error {
      // Check if entry exists
      b := tx.Bucket([]byte("alias")) 
      v := b.Get([]byte(alias_name))
      if v == nil {
          // If it does not exist, insert 
          b.Put([]byte(alias_name), []byte(target_name))
          reply := fmt.Sprintf(MSG_ADD_ALIAS_PASS, target_name, alias_name)
          ses.ChannelMessageSend(msg.ChannelID, reply)

      } else {
        // if it does, then there's a duplicate
        reply := fmt.Sprintf(MSG_ADD_ALIAS_DUPLICATE_FOUND, alias_name)
        ses.ChannelMessageSend(msg.ChannelID, reply) 
      }
      return nil
    })

  }

}


func cmd_resist(ses *discordgo.Session, msg *discordgo.MessageCreate, args []string) {
  if len(args) != 3 || args[1] != "vs" {
    reply := fmt.Sprintf(MSG_HELP_QUERY, MSG_RESIST_HELP)
    ses.ChannelMessageSend(msg.ChannelID, reply)
    return
  }

    Active, AtoiErr := strconv.Atoi(args[0])
    if AtoiErr != nil {
        fmt.Println("[cmd_resist] Itoa Error on Active")
    reply := fmt.Sprintf(MSG_HELP_QUERY, MSG_RESIST_HELP)
    ses.ChannelMessageSend(msg.ChannelID, reply)
        return
   }

    Passive, AtoiErr := strconv.Atoi(args[2]) 
    if AtoiErr != nil {
        fmt.Println("[cmd_resist] Itoa Error on Passive")
    reply := fmt.Sprintf(MSG_HELP_QUERY, MSG_RESIST_HELP)
    ses.ChannelMessageSend(msg.ChannelID, reply)
        return
    }

    Result := (Active - Passive) * 5 + 50
    reply := fmt.Sprintf(MSG_RESIST_THINK, Active, Passive)
    if Result > 95 {
        reply += MSG_RESIST_AUTO_PASS 
    } else if Result < 5 {
        reply += MSG_RESIST_AUTO_FAIL 
    } else {
        reply += fmt.Sprintf(MSG_RESIST_NORMAL, Result) 
    }
    ses.ChannelMessageSend(msg.ChannelID, reply)
}

