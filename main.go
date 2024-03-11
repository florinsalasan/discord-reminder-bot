package main

import (
    "discord-reminder-bot/bot"
    "github.com/joho/godotenv"
    "log"
    "os"
)

func main() {

    // Load .env with the godotenv module
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    // Get the token from .env
    botToken := os.Getenv("BOT_TOKEN")
    guildID := os.Getenv("GUILD_ID")

    // set the bot's token and start the bot
    bot.BotToken = botToken
    bot.GuildID = guildID
    bot.Run()

}
