package main

import (
    "os"
    "log"
    //"fmt"
    "github.com/bwmarrin/discordgo"
    "github.com/joho/godotenv"
)

func main() {

    err := godotenv.Load(".env")
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    botToken := os.Getenv("BOT_TOKEN")

    discord, err := discordgo.New("Bot " + botToken)

}
