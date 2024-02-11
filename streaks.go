package main

import (
    "fmt"
    "log"
    "os"
    "os/signal"
    "time"

    "github.com/joho/godotenv"

    "github.com/bwmarrin/discordgo"
)

func main() {
    dotenverr := godotenv.Load()
    if dotenverr != nil {
        log.Fatal("Error loading .env file")
    }

    APP_ID := os.Getenv("APP_ID")
    PUBLIC_KEY := os.Getenv("PUBLIC_KEY")
    DISCORD_TOKEN := os.Getenv("DISCORD_TOKEN")

    discord, _ := discordgo.New("Bot " + DISCORD_TOKEN)
    discord.AddHandler(func(discord *discordgo.Session, r *discordgo.Ready) {
        fmt.Println("Bot is ready")
    })

    err := discord.Open()
    if err != nil {
        log.Fatalf("Cannot open the session: %v", err)
    }
    defer discord.Close()

    event := createEvent(discord)

    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt)
    <-stop
    log.Println("Shutdown")

}

func createEvent(discord *discordgo.Session) {
}
