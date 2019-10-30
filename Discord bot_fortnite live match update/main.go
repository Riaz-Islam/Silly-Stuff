package main

import (
	"Bot/tracker"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"os"
	"strings"
	"sync"
	_ "unicode"
)

const myChannel = "594796465155735552"

const (
	status     = "&help"
	helpString = `Here's a list of commands:

*Fortnite commands:*
**&subscribe** type this in the channel where where you want to get updates of match (you will only get update of the people you registered)
**&register <epic username>** to get live match updates of anyone
**&showreg** shows the list of registered users
**&unsubscribe**
**&unregister <epic username>**

*report any bug to code6044@gmail.com*
`
)

var (
	commandPrefix uint8
	botID         string
)

func main() {
	//tracker.LoggerInit("./debug.log", 3600*24, 1024*1024*1024, 100, 3)
	tracker.GlobalStat = make(map[string]*tracker.ServerStat)
	tracker.Mutex = new(sync.Mutex)
	//err := tracker.CacheInit()
	//if err != nil {
	//	tracker.LogInfo("Cache init() error :", err.Error())
	//	panic(err)
	//}

	commandPrefix = '&'

	discord, err := discordgo.New("Bot NTk0NzkyNjMwNzE4NjkzMzc3.XRhp2g.x0LMm5dqYQehv9X4feoHlfZRHMc")
	errCheck("error creating discord session", err)
	user, err := discord.User("@me")
	errCheck("error retrieving account", err)

	botID = user.ID
	//tracker.LogInfo("botID:", botID)
	discord.AddHandler(commandHandler)
	discord.AddHandler(func(discord *discordgo.Session, ready *discordgo.Ready) {
		tracker.GlobalSession = discord
		err = discord.UpdateStatus(0, status)
		if err != nil {
			tracker.LogWarning("Error attempting to set my status")
		}
		servers := discord.State.Guilds
		fmt.Printf("Bolod has started on %d servers\n", len(servers))

		tracker.Mutex.Lock()
		//tracker.LogVerbosef("lock acquired by addhandler")
		for _, server := range servers {
			if _, err := os.Stat(tracker.DataDir + server.ID + tracker.Extension); os.IsNotExist(err) {
				tracker.LogInfof("id=%s, name=%s err=%v", server.ID, server.Name, err)
				tracker.GlobalStat[server.ID] = &tracker.ServerStat{}
				tracker.GlobalStat[server.ID].ServerId = server.ID
				tracker.GlobalStat[server.ID].ServerName = server.Name
				tracker.GlobalStat[server.ID].ChannelId = ""
				tracker.GlobalStat[server.ID].Stats = make([]tracker.Stat, 0)
				tracker.SetServerStats(server.ID, tracker.GlobalStat[server.ID])
			} else {
				tracker.GlobalStat[server.ID] = tracker.GetServerStats(server.ID)
			}
		}
		//tracker.LogVerbosef("lock released by addhandler")
		tracker.Mutex.Unlock()
		go tracker.Check(discord)
	})

	err = discord.Open()
	go tracker.PeriodicUpdates(discord)
	errCheck("Error opening connection to Discord", err)
	defer discord.Close()

	<-make(chan struct{})

}

func commandHandler(discord *discordgo.Session, message *discordgo.MessageCreate) {
	tracker.GlobalSession = discord
	server, _ := discord.State.Guild(message.GuildID)
	channel, _ := discord.State.Channel(message.ChannelID)
	var serverName, serverId string
	if server == nil {
		serverName = "DM"
		serverId = "0"
	} else {
		serverName = server.Name
		serverId = server.ID
	}

	if message.GuildID == "264445053596991498" {
		return
	}
	user := message.Author

	if user.ID != botID {
		tracker.LogInfo("Message from: ", user.Username, " in ", serverName+"("+serverId+"):"+channel.Name+"("+channel.ID+"):", "->", message.Content)
	}
	if user.ID == botID || user.Bot {
		//tracker.LogInfo("ignored message from bot")
		return
	}

	if len(message.GuildID) == 0 {
		discord.ChannelMessageSend(message.ChannelID, "ট্রিট দেন ভাই 😇")
		return
	}

	content := message.Content
	if len(content) == 0 || content[0] != '&' {
		return
	}
	list := strings.Split(content, " ")

	tracker.LogVerbose(content)
	command := strings.ToLower(list[0][1:])
	//if len(list) == 1 &&  command != "help" {
	//	exit(discord, message)
	//	return
	//}
	details := strings.Join(list[1:], " ")

	switch command {
	case "help":
		discord.ChannelMessageSend(message.ChannelID, helpString)
	case "register":
		tracker.LogInfo("Registering - ", details)
		err := tracker.Register(details, message.GuildID)
		if err != nil {
			discord.ChannelMessageSend(message.ChannelID, err.Error())
		} else {
			discord.ChannelMessageSend(message.ChannelID, "🤗 🤗 🤗")
		}
	case "unregister":
		_ = tracker.Unregister(details, message.GuildID)
		discord.ChannelMessageSend(message.ChannelID, "😱 😥 😰")
	case "showreg":
		rep, err := tracker.Show(message.GuildID)
		if err != nil {
			exit(discord, message)
		} else {
			discord.ChannelMessageSend(message.ChannelID, rep)
		}
	case "subscribe":
		channel := message.ChannelID
		server := message.GuildID
		tracker.SetChannel(channel, server)
		discord.ChannelMessageSend(message.ChannelID, "👍")
	case "unsubscribe":
		server := message.GuildID
		tracker.UnSetChannel(server)
		discord.ChannelMessageSend(message.ChannelID, "😭 😭 😭")
	case "2360":
		tracker.BroadCast(discord, details)
	default:
		exit(discord, message)
	}
	tracker.LogInfo("executed command: ", command)
}

func mention(user *discordgo.User) string {
	return "<@" + user.ID + ">"
}

func errCheck(msg string, err error) {
	if err != nil {
		fmt.Printf("%s: %+v", msg, err)
		panic(err)
	}
}

func exit(discord *discordgo.Session, message *discordgo.MessageCreate) {
	//discord.ChannelMessageSend(message.ChannelID, `Your command format is not recognized <@` + message.Author.ID +  `> 😢`)
}
