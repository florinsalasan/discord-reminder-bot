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
        {
            // TODO: Fix the event handler for add-topic, as it will be similar for
            // other commands later on.
            // Do not touch this for now, the command is being shown properly on 
            // discord, however I am not handling the input correctly yet
            Name: "add-topic",
            // use the subcommands usage to implement the frequency of the reminders
            Description: "Parent command for adding a topic to be reminded of, options set the frequency",
            Options: []*discordgo.ApplicationCommandOption {
                {
                    Type: discordgo.ApplicationCommandOptionString,
                    Name: "frequency",
                    Description: "Set the new reminder frequency to user-input",
                    Required: true,
                    Choices: []*discordgo.ApplicationCommandOptionChoice{
                        {
                            Name: "Daily",
                            Value: "daily",
                        },
                        {
                            Name: "Weekly",
                            Value: "weekly",
                        },
                        {
                            Name: "Monthly",
                            Value: "monthly",
                        },
                        {
                            Name: "Yearly",
                            Value: "yearly",
                        },
                    },
                },
                {
                    Type: discordgo.ApplicationCommandOptionString,
                    Name: "topic",
                    Description: "The topic you want to be reminded of",
                    Required: true,
                },
            },
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
        "add-topic": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
            // Can access options in the order given, or if we wanted to could have
            // converted this into a map
            options := i.ApplicationCommandData().Options
            content := ""
            frequency := ""
            topic := options[1].StringValue()

            // This is how to get the values that the user specifies:
            // println(options[0].StringValue())
            // println(options[1].StringValue())
            switch options[0].StringValue() {
            // Swap this out for a function that sets the reminders with frequency
            // as a parameter, zero need for a switch statement here.
            case "daily":
                frequency = "daily"
                addReminderTopics(s, topic, frequency)
            default:
                content = "Sorry only daily frequencies have been implemented so far"
            }

            if frequency != "daily" {
                log.Panic("Frequencies other than daily have not been implemented yet, sorry")
            }

            content = topic + " has been registered to receive " + 
                frequency + " updates."

            // At this point we have built the Content message that the bot willl 
            // respond with, need:
            // TODO:
            // Modify the pinned comment in the reminder-topics channel based on the 

            s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
                Type: discordgo.InteractionResponseChannelMessageWithSource,
                Data: &discordgo.InteractionResponseData{
                    Content: content,
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
            log.Panicf("Cannot create '%v' command: %v", v.Name, err)
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

    // Check that only one pinned message exists to read from
    if len(reminders) != 1 {
        log.Fatal("More than one pinned message to read the reminders from")
    }

    // Split the one message into the different 
    topics := strings.Split(reminders[0].Content, ",")

    // trim the whitespace for consistency
    for i, topic := range topics {
        topics[i] = strings.TrimSpace(topic)
    }

    for _, topic := range topics {
        println(topic)
    }

}

func addReminderTopics(discord *discordgo.Session, topic string, freq string) {

    // Currently freq will do nothing, need to think of the best way to implement
    // frequencies other than daily

    reminders, err := discord.ChannelMessagesPinned(ReminderChannelID)
    if err != nil {
        log.Fatal("Couldn't get the list of reminders to remind user of")
    }

    // Check that only one pinned message exists to read from
    if len(reminders) != 1 {
        log.Fatal("More than one pinned message to read the reminders from")
    }

    messageID := reminders[0].ID

    // Split the one message into the different 
    topics := strings.Split(reminders[0].Content, ",")

    // trim the whitespace for consistency
    for i, topic := range topics {
        topics[i] = strings.TrimSpace(topic)
    }

    for _, topic := range topics {
        println(topic)
    }

    topics = append(topics, topic)
    topicsStringed := strings.Join(topics[:], ", ")
    
    // rejoined the list of topics into a string again, and now need to
    // edit the pinned message to the new string. Ok so bots are seemingly not 
    // allowed to modify messages that were not sent by it  
    // So need to remove the initial pinned message, pin the inital value and 
    // afterwards can continue modifying the pinned comment
    msg, err := discord.ChannelMessageEdit(ReminderChannelID, messageID, topicsStringed)
    if err != nil {
        log.Panicf("Cannot modify pinned message: %v", err)
    }

    println(msg.Content)

}

