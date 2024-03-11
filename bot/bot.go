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
    RemoveCommands bool
    GuildID string
    ReminderChannelID string
)

var (

    commands = []*discordgo.ApplicationCommand{
        {
            Name: "test-command",
            // Commands and options must have descriptions, if a command
            // or option does not have one, it will not be registered.
            Description: "Meant to test the slash commands working",
        },
    }

    commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        "test-command": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
            s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
                Type: discordgo.InteractionResponseChannelMessageWithSource,
                Data: &discordgo.InteractionResponseData{
                    Content: "Hey there, congrats on finding the first slash cmd!",
                },
            })
        },
    }
)

func Run() {

    discord, err := discordgo.New("Bot " + BotToken)
    if err != nil {
        log.Fatal(err)
    }

    // Add an event handler, with the handler function of newMessage
    discord.AddHandler(newMessage)
    discord.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
            h(s, i)
        }
    })

    // open the discord session and defer it's closing
    err = discord.Open()
    if err != nil {
        log.Fatal("error opening connection, ", err)
    }
    defer discord.Close()

    // Get the reminder topics from the channel
    getReminderTopics(discord)

    // Add in the commands that were defined earlier.
    registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
    for i, v := range commands {
        cmd, err := discord.ApplicationCommandCreate(discord.State.User.ID, GuildID, v)
        if err != nil {
            log.Panicf("Cannot create '%v' comand: %v", v.Name, err)
        }
        registeredCommands[i] = cmd
    }

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

func getReminderTopics(discord *discordgo.Session) {

    // Get the most recent message in reminder-topics channel, ID is in
    // ReminderChannelID and we do this by calling ChannelMessagesPinned on 
    // the current session.
    reminders, err := discord.ChannelMessagesPinned(ReminderChannelID)
    if err != nil {
        log.Fatal("Couldn't get the list of reminders to remind user of")
    }

    println(len(reminders))
    for _, rem := range reminders {
        println(rem)
    }

}
