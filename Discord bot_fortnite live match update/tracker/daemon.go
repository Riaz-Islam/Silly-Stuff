package tracker

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"time"
)

var statusChannel = "596163118921678868"
var lastLifeCheck = time.Now()

func BotStatPrint(discord *discordgo.Session) {
	servers := discord.State.Guilds
	message := fmt.Sprintf("**Bot is alive on %d servers**\n", len(servers))
	for _, s := range servers {
		message += fmt.Sprintln("**Server:", s.Name, "**\nRegion:", s.Region, "\nMembers:", len(s.Members), "\nJoined:", s.JoinedAt, "\n")
	}
	discord.ChannelMessageSend(statusChannel, message)
}

func PeriodicUpdates(discord *discordgo.Session) {
	GlobalSession = discord
	for {
		//if (time.Now().Hour() == 21 || time.Now().Hour() == 12) && time.Now().Minute() == 5 {
		//	PrintNews("", discord)
		//	time.Sleep(2 * time.Minute)
		//}

		if time.Now().Sub(lastLifeCheck) > 20*time.Minute {
			BotStatPrint(discord)
			lastLifeCheck = time.Now()
		}
	}
}
