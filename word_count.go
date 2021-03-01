package main

import (
	"github.com/bwmarrin/discordgo"

	"regexp"
	"time"
)

type Word_Count_Entry struct {
	Word    string
	Count   int
	Regex   *regexp.Regexp
	Created int64
}

// TODO: Maybe we shift this to a better form of offline DB?
var GlobalWordCountDatabase map[string]*Word_Count_Entry

const WordCountTimeLimit int64 = 43200 // 12 hours in seconds

func SchedCleanWordCountDatabase() {
	for K, V := range GlobalWordCountDatabase {
		Diff := time.Now().Unix() - V.Created
		if Diff > WordCountTimeLimit {
			delete(GlobalWordCountDatabase, K)
		}
	}
}

func InitWordCount() {
	GlobalWordCountDatabase = make(map[string]*Word_Count_Entry)
}

func UpdateWordCount(Msg *discordgo.MessageCreate) {
	// Check if channel is being tracked
	Entry, Exists := GlobalWordCountDatabase[Msg.ChannelID]
	if Exists {
		Strs := Entry.Regex.FindAllString(Msg.Content, -1)
		if Strs != nil {
			Count := len(Strs)
			Entry.Count += Count
		}
	}
}
