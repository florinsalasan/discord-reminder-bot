package bot

import (
    "fmt"
    "log"
    "os"
    "os/signal"
    "strings"

    "github.com/bwmarrin/discordgo"
)

var (
    BotToken string
)

func Run() {

    discord, err := discordgo.New("Bot " + BotToken)
    if err != nil {
        log.Fatal(err)
    }

    // Add an event handler, with the handler function of newMessage
    discord.AddHandler(newMessage)

    // open the discord session and defer it's closing
    discord.Open()
    defer discord.Close()

    // This section will run until the process is terminated
    fmt.Println("Bot running...")
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    <-c

}

func newMessage(discord *discordgo.Session, message *discordgo.MessageCreate) {

    // Ignore the bot messages
    if message.Author.ID == discord.State.User.ID {
        return
    }

    // handle the different messages sent by a user
    switch {
    case strings.Contains(message.Content, "reminder"):
        discord.ChannelMessageSend(message.ChannelID, "Will remind you!")
    case strings.Contains(message.Content, "bot"):
        discord.ChannelMessageSend(message.ChannelID, "Hello from reminder bot!")
    }

}
