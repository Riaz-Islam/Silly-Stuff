package tracker

import (
	"log"
	"os"
)

func InitBotLog(BotLog *log.Logger) *os.File {
	Blog, err := os.Create("./Bot.Log")
	if err != nil {
		log.Fatalf("error opening tcp log file: %v", err)
	}
	BotLog = log.New(Blog, "BOT: ", log.Ldate|log.Ltime|log.Lshortfile)
	return  Blog
}
