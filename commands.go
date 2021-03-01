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
	GlobalCommandArgSplitter *regexp.Regexp
)

const DiscordMessageMaxChars int = 2000

func InitCommands() {
	GlobalCommandArgSplitter = regexp.MustCompile(`(?i)(?:[^\s"]+\b|:|(")[^"]*("))+|[=!&|~+\-\*\/\%]`)
	if GlobalCommandArgSplitter == nil {
		Panik("commandArgSplitter failed to compile")
	}
}

func ProcCommands(Ses *discordgo.Session, Msg *discordgo.MessageCreate) {
	Args := GlobalCommandArgSplitter.FindAllString(Msg.Content, -1)
	if Args != nil && len(Args) >= 2 {
		CommandStr := Args[1]
		Args = Args[2:] // get array from 2 to n

		// Use this instead of interfaces
		// Clearer, more concise, easier to debug
		switch CommandStr {
        case "help":
            Ses.ChannelMessageSend(Msg.ChannelID, MsgHelp)
		case "version":
			Ses.ChannelMessageSend(Msg.ChannelID, MsgVersion)
        case "add-alias":
            CmdAddAlias(Ses, Msg, Args)
        case "get-alias":
            CmdGetAlias(Ses, Msg, Args)
        case "remove-alias":
            CmdRemoveAlias(Ses, Msg, Args)
        case "find":
            CmdFind(Ses, Msg, Args)
		case "resist":
            CmdResist(Ses, Msg, Args)
	    }
    }
}

func WrapCode(Str string) string {
	return "```" + Str + "```"
}

func CmdFind(Ses *discordgo.Session, Msg *discordgo.MessageCreate, Args []string) {
	if len(Args) != 1 {
		Reply := fmt.Sprintf(MsgHelpQuery, MsgFindHelp)
		Ses.ChannelMessageSend(Msg.ChannelID, Reply)
		return
	}
	//defer recovery(discord, message)

	Key := strings.ToLower(Args[0])

	// Check if it is an alias. If so, get the actual key from alias.
    GlobalDb.View(func(Tx *bolt.Tx) error {
        B := Tx.Bucket([]byte("alias")) 
        AliasValue := B.Get([]byte(Key))
        if AliasValue != nil {
            // Key is an alias 
            // Thus, find the entry from dictionary using alias value
            AliasValueStr := string(AliasValue)
            Entry, Exist := GlobalDictionary[AliasValueStr]
            if !Exist {
                Ses.ChannelMessageSend(Msg.ChannelID, MsgFindFail)
                return nil
            }
            EntryStr := string(Entry)
            Reply := fmt.Sprintf(MsgFindFoundWithAlias, Key, AliasValueStr, EntryStr) 
            Ses.ChannelMessageSend(Msg.ChannelID, Reply)
            return nil
        } 
       
        // Key is an alias 
        // Thus, find the entry from dictionary using alias value
        Entry, Exist := GlobalDictionary[Key]
        if !Exist {
            Ses.ChannelMessageSend(Msg.ChannelID, MsgFindFail)
            return nil
        }
        EntryStr := string(Entry)
        Reply := fmt.Sprintf(MsgFindFound, Key, EntryStr) 
        Ses.ChannelMessageSend(Msg.ChannelID, Reply)
        return nil
    })


}

func CmdRemoveAlias(Ses *discordgo.Session, Msg *discordgo.MessageCreate, Args []string) {
	if len(Args) <= 0 {
		Reply := fmt.Sprintf(MsgHelpQuery, MsgRemoveAliasHelp)
		Ses.ChannelMessageSend(Msg.ChannelID, Reply)
		return
	}

	AliasName := strings.Join(Args[0:], " ")
    GlobalDb.Update(func(Tx *bolt.Tx) error {
        // Check if entry exists
        B := Tx.Bucket([]byte("alias")) 
        V := B.Get([]byte(AliasName))
        if V == nil {
            // If it does not exist, we failed  
            Reply := fmt.Sprintf(MsgRemoveAliasFailure, AliasName)
            Ses.ChannelMessageSend(Msg.ChannelID, Reply)

        } else {
            // if it does, then there's a duplicate
            B.Delete([]byte(AliasName))
            Reply := fmt.Sprintf(MsgRemoveAliasSuccess, AliasName)
            Ses.ChannelMessageSend(Msg.ChannelID, Reply) 
        }
        return nil
    })
}

func CmdGetAlias(Ses *discordgo.Session, Msg *discordgo.MessageCreate, Args []string) {
	if len(Args) <= 0 {
		Reply := fmt.Sprintf(MsgHelpQuery, MsgGetAliasHelp)
		Ses.ChannelMessageSend(Msg.ChannelID, Reply)
		return
	}

	AliasName := strings.Join(Args[0:], " ")

    // Select value from alias where key = ?
    GlobalDb.View(func(Tx *bolt.Tx) error {
        // Check if entry exists
        B := Tx.Bucket([]byte("alias")) 
        TargetName := B.Get([]byte(AliasName))
        if TargetName == nil {
            // Does not exist
            Reply := fmt.Sprintf(MsgGetAliasFailed, AliasName)
            Ses.ChannelMessageSend(Msg.ChannelID, Reply)
        } else {
            // If it does, then there's a duplicate
            Reply := fmt.Sprintf(MsgGetAliasSuccess, TargetName, AliasName)
            Ses.ChannelMessageSend(Msg.ChannelID, Reply) 
        }
        return nil
    })
}

func CmdAddAlias(Ses *discordgo.Session, Msg *discordgo.MessageCreate, Args []string) {
    if len(Args) <= 0 {
		Reply := fmt.Sprintf(MsgHelpQuery, MsgAddAliasHelp)
		Ses.ChannelMessageSend(Msg.ChannelID, Reply)
		return
    }
    
    // combine '<alias> = <target>' into one string
	StrToParse := strings.Join(Args[0:], " ")           	

    // split to: '<alias>', '=', '<target>'
    StrToParseArray := strings.Split(StrToParse, " = ") 	
    if len(StrToParseArray) != 2 {                      
        // we expect 2 items in the array: '<alias>' and '<target>'
		Reply := fmt.Sprintf(MsgHelpQuery, MsgAddAliasHelp)
		Ses.ChannelMessageSend(Msg.ChannelID, Reply)
		return
	}
	AliasName := strings.ToLower(StrToParseArray[0])
	TargetName := strings.ToLower(StrToParseArray[1])
    
    if _, Exist := GlobalDictionary[TargetName]; !Exist {
		Ses.ChannelMessageSend(Msg.ChannelID, MsgAddAliasTargetNotFound)
    } else {
        // TODO: We can optimize by using View to check 
        // if an alias exist first, then use Update if it doesn't.
        // Then again, it might be safer to just lock the whole thing.
        // Whatever #lazyprogramming
        GlobalDb.Update(func(Tx *bolt.Tx) error {
            // Check if entry exists
            B := Tx.Bucket([]byte("alias")) 
            V := B.Get([]byte(AliasName))
            if V == nil {
                // If it does not exist, insert 
                B.Put([]byte(AliasName), []byte(TargetName))
                Reply := fmt.Sprintf(MsgAddAliasSuccess, TargetName, AliasName)
                Ses.ChannelMessageSend(Msg.ChannelID, Reply)

            } else {
                // if it does, then there's a duplicate
                Reply := fmt.Sprintf(MsgAddAliasDuplicateFound, AliasName)
                Ses.ChannelMessageSend(Msg.ChannelID, Reply) 
            }
            return nil
        })

    }

}


func CmdResist(Ses *discordgo.Session, Msg *discordgo.MessageCreate, Args []string) {
	if len(Args) != 3 || Args[1] != "vs" {
		Reply := fmt.Sprintf(MsgHelpQuery, MsgResistHelp)
		Ses.ChannelMessageSend(Msg.ChannelID, Reply)
		return
	}

    Active, AtoiErr := strconv.Atoi(Args[0])
    if AtoiErr != nil {
        fmt.Println("[CmdResist] Itoa Error on Active")
		Reply := fmt.Sprintf(MsgHelpQuery, MsgResistHelp)
		Ses.ChannelMessageSend(Msg.ChannelID, Reply)
        return
   }

    Passive, AtoiErr := strconv.Atoi(Args[2]) 
    if AtoiErr != nil {
        fmt.Println("[CmdResist] Itoa Error on Passive")
		Reply := fmt.Sprintf(MsgHelpQuery, MsgResistHelp)
		Ses.ChannelMessageSend(Msg.ChannelID, Reply)
        return
    }

    Result := (Active - Passive) * 5 + 50
    Reply := fmt.Sprintf(MsgResistThink, Active, Passive)
    if Result > 95 {
        Reply += MsgResistAutoSuccess 
    } else if Result < 5 {
        Reply += MsgResistAutoFail 
    } else {
        Reply += fmt.Sprintf(MsgResistNormal, Result) 
    }
    Ses.ChannelMessageSend(Msg.ChannelID, Reply)
}

