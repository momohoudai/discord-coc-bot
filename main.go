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
	GlobalBotId       string
	GlobalArgSplitter *regexp.Regexp
    GlobalDictionary  map[string]string
    GlobalDb     *bolt.DB 
)

func MessageHandler(Ses *discordgo.Session, Msg *discordgo.MessageCreate) {
	User := Msg.Author
	if User.ID == GlobalBotId || User.Bot {
		return
	}

	// For message parsing
	UpdateWordCount(Msg)

	// handle prefix
	Content := strings.ToLower(Msg.Content)
	if strings.HasPrefix(Content, "!coc") {
		go func() {
			defer Kalm(Ses, Msg, "ProcCommands")
			ProcCommands(Ses, Msg)
		}()
	} 
}

func ReadyHandler(Ses *discordgo.Session, Ready *discordgo.Ready) {
	Err := Ses.UpdateListeningStatus("'!coc help'")
	if Err != nil {
		fmt.Println("Error attempting to set my status")
	}
	servers := Ses.State.Guilds
	fmt.Printf("CocBot has started on %d servers\n", len(servers))
}

func Panik(Format string, a ...interface{}) {
	panic(fmt.Sprintf(Format, a...))
}

func Kalm(Ses *discordgo.Session, Msg *discordgo.MessageCreate, Name string) {
	if R := recover(); R != nil {
		fmt.Printf("[%s] Recovered: %v\n", Name, R)
		Ses.ChannelMessageSend(Msg.ChannelID, MsgGenericFail)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	Token, ReadFileErr := ioutil.ReadFile("TOKEN")
	if ReadFileErr != nil {
		Panik("Cannot read or find TOKEN file\n")
	}

	InitCommands()
    
    // Configure Dictionary
    JsonFile, OpenErr := os.Open("data.json")
	if OpenErr != nil {
        Panik("Cannot open data.json")
	}
	defer JsonFile.Close()

    JsonBytes, ReadAllErr := ioutil.ReadAll(JsonFile)
    if ReadAllErr != nil {
        Panik("Cannot read data.json");
    }

    UnmarshalErr := json.Unmarshal(JsonBytes, &GlobalDictionary)
    if UnmarshalErr != nil {
        Panik("Cannot unmarshal data.json")
    }

    // Configure Alias DB
    var DbOpenErr error
    GlobalDb, DbOpenErr = bolt.Open("./db", 0666, nil)
    if DbOpenErr != nil {
        Panik("Cannot open database: %s", DbOpenErr)
    }
    defer GlobalDb.Close()
    GlobalDb.Update(func (Tx *bolt.Tx) error {
        _, Err := Tx.CreateBucketIfNotExists([]byte("alias"))
        if Err != nil {
            return fmt.Errorf("Cannot create 'alias' bucket")
        }
        return nil
    })

	// Configure Discord
	Discord, Err := discordgo.New("Bot " + string(Token))
	if Err != nil {
		Panik("Cannot initialize discord: %s\n", Err.Error())
	}
	User, Err := Discord.User("@me")
	if Err != nil {
		Panik("Error retrieving account: %s\n", Err.Error())
	}
	GlobalBotId = User.ID
	Discord.AddHandler(MessageHandler)
	Discord.AddHandler(ReadyHandler)

	Err = Discord.Open()
	if Err != nil {
		Panik("Error retrieving account: %s\n", Err.Error())
	}
	defer Discord.Close()

	<-make(chan struct{})

}
