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

    // set the bot's token and start the bot
    bot.BotToken = botToken
    bot.Run()

}
