package main

import (
  "github.com/bwmarrin/discordgo"
  bolt "go.etcd.io/bbolt"
    
  "os"
  "encoding/json"
  "fmt"
  "io/ioutil"
  "math/rand"
  "regexp"
  "strings"
  "time"
)

var (
  global_bot_id        string
  global_arg_splitter  *regexp.Regexp
  global_dictionary    map[string]string
  global_db            *bolt.DB 
)

func handle_message(ses *discordgo.Session, msg *discordgo.MessageCreate) {
  user := msg.Author
  if user.ID == global_bot_id || user.Bot {
    return
  }

  // handle prefix
  content := strings.ToLower(msg.Content)
  if strings.HasPrefix(content, "!coc") {
    go func() {
      defer kalm(ses, msg, "execute_commands")
      execute_commands(ses, msg)
    }()
  } 
}

func handle_ready(ses *discordgo.Session, ready *discordgo.Ready) {
  err := ses.UpdateListeningStatus("'!coc help'")
  if err != nil {
    fmt.Println("Error attempting to set my status")
  }
  servers := ses.State.Guilds
  fmt.Printf("CocBot has started on %d servers\n", len(servers))
}

func panik(format string, a ...interface{}) {
  panic(fmt.Sprintf(format, a...))
}

func kalm(ses *discordgo.Session, msg *discordgo.MessageCreate, name string) {
  if r := recover(); r != nil {
    fmt.Printf("[%s] Recovered: %v\n", name, r)
    ses.ChannelMessageSend(msg.ChannelID, MSG_GENERIC_FAIL)
  }
}

func main() {
  rand.Seed(time.Now().UnixNano())
  token, read_file_err := ioutil.ReadFile("TOKEN")
  if read_file_err != nil {
    panik("Cannot read or find TOKEN file\n")
  }

  init_commands()

  // Configure Dictionary
  json_file, open_err := os.Open("data.json")
  if open_err != nil {
    panik("Cannot open data.json")
  }
  defer json_file.Close()

  json_bytes, read_all_err := ioutil.ReadAll(json_file)
  if read_all_err != nil {
    panik("Cannot read data.json");
  }

  unmarshal_err := json.Unmarshal(json_bytes, &global_dictionary)
  if unmarshal_err != nil {
    panik("Cannot unmarshal data.json")
  }

  // Configure Alias DB
  var db_open_err error
  global_db, db_open_err = bolt.Open("./db", 0666, nil)
  if db_open_err != nil {
    panik("Cannot open database: %s", db_open_err)
  }
  defer global_db.Close()
  global_db.Update(func (tx *bolt.Tx) error {
      _, err := tx.CreateBucketIfNotExists([]byte("alias"))
      if err != nil {
          return fmt.Errorf("Cannot create 'alias' bucket")
      }
      return nil
  })

  // Configure Discord
  discord, err := discordgo.New("Bot " + string(token))
  if err != nil {
    panik("Cannot initialize discord: %s\n", err.Error())
  }
  user, err := discord.User("@me")
  if err != nil {
    panik("Error retrieving account: %s\n", err.Error())
  }
  global_bot_id = user.ID
  discord.AddHandler(handle_message)
  discord.AddHandler(handle_ready)

  err = discord.Open()
  if err != nil {
    panik("Error retrieving account: %s\n", err.Error())
  }
  defer discord.Close()  
  <-make(chan struct{})

}
