package tracker

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
)

const (
	INFOCHAN    = "638951837654843392"
	WARNINGCHAN = "638951895779508224"
	VERBOSECHAN = "638951968928301074"
)

func GetRandomTextChannel(serverID string, discord *discordgo.Session) *discordgo.Channel {
	server, _ := discord.State.Guild(serverID)
	for _, c := range server.Channels {
		if c.Type == 0 {
			return c
		}
	}
	return nil
}

func Highlight(message, theme string) string {
	ret := "```" + theme + "\n" + message + "\n```\n"
	return ret
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Log(channel string, pre string, m ...interface{}) {
	message := pre + " " + fmt.Sprint(m...)
	GlobalSession.ChannelMessageSend(channel, message)
}

func Logf(channel, pre, a string, m ...interface{}) {
	message := pre + " " + fmt.Sprintf(a, m...)
	GlobalSession.ChannelMessageSend(channel, message)
}

func LogInfo(m ...interface{}) {
	pre := ""
	Log(INFOCHAN, pre, m...)
}

func LogInfof(a string, m ...interface{}) {
	pre := ""
	Logf(INFOCHAN, pre, a, m...)
}

func LogWarning(m ...interface{}) {
	pre := ""
	Log(WARNINGCHAN, pre, m...)
}

func LogWarningf(a string, m ...interface{}) {
	pre := ""
	Logf(WARNINGCHAN, pre, a, m...)
}

func LogVerbose(m ...interface{}) {
	pre := ""
	Log(VERBOSECHAN, pre, m...)
}

func LogVerbosef(a string, m ...interface{}) {
	pre := ""
	Logf(VERBOSECHAN, pre, a, m...)
}
