package main

import (
    "fmt"
    "log"
    "os"
    "os/signal"
    "time"
    "strings"
    "errors"

    "github.com/joho/godotenv"

    "github.com/bwmarrin/discordgo"
)

var discSess *discordgo.Session

func main() {
    dotenverr := godotenv.Load(".env")
    if dotenverr != nil {
        log.Fatal("Error loading .env file")
    }

    APP_ID := os.Getenv("APP_ID")
    PUBLIC_KEY := os.Getenv("PUBLIC_KEY")
    DISCORD_TOKEN := os.Getenv("DISCORD_TOKEN")
    CHANNEL_ID := os.Getenv("CHANNEL_ID")
    GUILD_ID := os.Getenv("GUILD_ID")
    RemoveCommands := true

    discord, _ := discordgo.New("Bot " + DISCORD_TOKEN)
    discord.AddHandler(func(discord *discordgo.Session, r *discordgo.Ready) {
        fmt.Println("Bot is ready")
    })

    err := discord.Open()
    if err != nil {
        log.Fatalf("Cannot open the session: %v", err)
    }
    defer discord.Close()

    // event := createEvent(discord)

    fmt.Printf("APP_ID: %v\n", APP_ID)
    fmt.Printf("PUBLIC_KEY: %v\n", PUBLIC_KEY)
    fmt.Printf("DISCORD_TOKEN: %v\n", DISCORD_TOKEN)
    fmt.Printf("CHANNEL_ID: %v\n", CHANNEL_ID)
    fmt.Printf("GUILD_ID: %v\n", GUILD_ID)

    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt)
    <-stop
    log.Println("Shutdown")

}
