package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
    "time"

	"github.com/bwmarrin/discordgo"
)

// Define some global vars
var (
    BotToken string
    GuildID string
    ReminderChannelID string
    pinnedMessage *discordgo.Message
    dailies map[string]bool
    jsonFile *os.File
    AppID string
    MainChannelID string
)

// Define the commands and their handlers
var (
    commands = []*discordgo.ApplicationCommand{
        {
            Name: "test-command",
            // Commands and options must have descriptions, if a command
            // or option does not have one, it will not be registered.
            Description: "Meant to test the slash commands working",
        },
        {
            Name: "add-reminder-time",
            Description: "Add a new reminder time to the schedule",
            Options: []*discordgo.ApplicationCommandOption{
                {
                    Type: discordgo.ApplicationCommandOptionInteger,
                    Name: "hour",
                    Description: "Hour of the reminder (0-23)",
                    Required: true,
                },
            },
        },
        {
            Name: "remove-topic",
            // use the subcommands usage to implement the frequency of the reminders
            Description: "Parent command for removing a topic to be reminded of",
            Options: []*discordgo.ApplicationCommandOption {
                {
                    Type: discordgo.ApplicationCommandOptionString,
                    Name: "topic",
                    Description: "The topic you want to be reminded of",
                    Required: true,
                    Choices: getAllTopics(),
                },
            },
        },
        {
            Name: "add-topic",
            Description: "Parent command for adding a topic to be reminded of, options set the frequency",
            Options: []*discordgo.ApplicationCommandOption {
                {
                    Type: discordgo.ApplicationCommandOptionString,
                    Name: "topic",
                    Description: "The topic you want to be reminded of",
                    Required: true,
                },
            },
        },
        {
            Name: "finished",
            Description: "Use this to mark a daily topic as finished to prevent more reminders for the rest of the day",
            Options: []*discordgo.ApplicationCommandOption {
                {
                    Type: discordgo.ApplicationCommandOptionString,
                    Name: "topic",
                    Description: "The topic you finished today",
                    Required: true,
                    Choices: getUnfinishedTopics(),
                },
            },
        },
        {
            Name: "get-unfinished",
            Description: "Return a list of topics that have not yet been marked as finished",
        },
        {
            Name: "reset",
            Description: "Use this to reset all topics to incomplete",
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

            options := i.ApplicationCommandData().Options
            content := ""
            topic := options[0].StringValue()
            content = topic + " has been registered to receive daily reminders."

            updateReminderTopic(s, topic, true)

            s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
                Type: discordgo.InteractionResponseChannelMessageWithSource,
                Data: &discordgo.InteractionResponseData{
                    Content: content,
                },
            })
        },
        "remove-topic": func(s *discordgo.Session, i *discordgo.InteractionCreate) {

            msgs, err := s.ChannelMessagesPinned(ReminderChannelID)
            if err != nil {
                log.Fatal("Something went wrong retrieving the pinned message when trying to remove-topic ", err)
            }
            if len(msgs) == 0 {
                s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
                    Type: discordgo.InteractionResponseChannelMessageWithSource,
                    Data: &discordgo.InteractionResponseData{
                        Content: "No topics to remove from",
                    },
                })
                return
            }

            options := i.ApplicationCommandData().Options
            topic := options[0].StringValue()

            updateReminderTopic(s, topic, false)

            s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
                Type: discordgo.InteractionResponseChannelMessageWithSource,
                Data: &discordgo.InteractionResponseData{
                    Content: topic + " has been removed from list of reminder topics",
                },
            })
        },
        "finished": func(s *discordgo.Session, i *discordgo.InteractionCreate) {

            topic := i.ApplicationCommandData().Options[0].StringValue()

            markedFinished := markDailyCompleted(topic)

            if markedFinished == true {
                s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
                    Type: discordgo.InteractionResponseChannelMessageWithSource,
                    Data: &discordgo.InteractionResponseData{
                        Content: topic + " has been marked as finished",
                    },
                })
                err := updateFinishedCommand(s)
                if err != nil {
                    log.Printf("Error updating finished command: %v", err)
                }
            } else {
                // Discord itself does not allow for choices that are not in the list to be used
                // So this section is never run.
                s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
                    Type: discordgo.InteractionResponseChannelMessageWithSource,
                    Data: &discordgo.InteractionResponseData{
                        Content: topic + " is not in the list, could not mark as finished",
                    },
                })
            }
        },
        "get-unfinished": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
            // get a list of topics that still need to be completed
            unfinished := getUnfinishedTopics()
            var topicCollection []string
            
            for topic := range unfinished {
                topicCollection = append(topicCollection, unfinished[topic].Name)
            }
                
        
            s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
                Type: discordgo.InteractionResponseChannelMessageWithSource,
                Data: &discordgo.InteractionResponseData{
                    Content: strings.Join(topicCollection, "\n"),
                },
            })
        },
        "reset": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
            // reset all of the topics to false once more.
            resetReminders()

            s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
                Type: discordgo.InteractionResponseChannelMessageWithSource,
                Data: &discordgo.InteractionResponseData{
                    Content: "Set all of the topics to incomplete!",
                },
            })
            err := updateFinishedCommand(s)
            if err != nil {
                log.Printf("Error updating finished command: %v", err)
            }
        },
    }
)

// #####################################################################
// FUNCTIONS BEGIN HERE ################################################
// #####################################################################

// think main instead of Run
func Run() {

    discord, err := discordgo.New("Bot " + BotToken)
    if err != nil {
        log.Fatal(err)
    }

    // open the discord session and defer it's closing
    err = discord.Open()
    if err != nil {
        log.Fatal("error opening connection, ", err)
    }
    defer discord.Close()

    // open the json file to fill out dailies before setting handlers
    jsonFile, err = os.Open("reminders.json")
    if err != nil {
        log.Fatal("couldn't open reminders.json ", err)
    }

    defer jsonFile.Close()

    // read the json file
    byteValue, _ := io.ReadAll(jsonFile)
    if len(byteValue) != 0 {
        json.Unmarshal(byteValue, &dailies)
        println("Found written values in the json file")
    } else {
        dailies = make(map[string]bool)
        println("did not find written values in the json file")
    }

    // Add the event handlers
    discord.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
            h(s, i)
        }
    })

    // Add in the commands that were defined earlier.
    registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
    for i, v := range commands {
        cmd, err := discord.ApplicationCommandCreate(discord.State.User.ID, GuildID, v)
        if err != nil {
            log.Panicf("Cannot create '%v' command: %v", v.Name, err)
        }
        registeredCommands[i] = cmd
    }

    // Get the pinned message containing the reminder topics
    messages, err := discord.ChannelMessagesPinned(ReminderChannelID)
    if err != nil {
        log.Fatal("Errored getting the pinned message from reminder channel when launching the bot")
    }

    if len(messages) == 0 {
        // create an initial message and pin it.
        message, err := discord.ChannelMessageSend(ReminderChannelID, "This is where your reminder topics will be stored!")
        if err != nil {
            log.Fatal("error sending initial pinned message to modify")
        }
        // Pin the message
        err = discord.ChannelMessagePin(ReminderChannelID, message.ID)
        if err != nil {
            log.Fatal("error pinning the initial message")
        }

        pinnedMessage = message

    }
    
    if len(messages) != 0 {
        pinnedMessage = messages[0]
    }

    startScheduledTasks(discord)

    // This section will run until the process is terminated
    fmt.Println("Bot running...")
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    <-c

}

// HELPER FUNCTIONS
func updateReminderTopic(discord *discordgo.Session, topic string, add bool) {
    // helper to update reminder topic pinned message, adds a topic if add is true
    // removes topic otherwise.
    
    // Use the pinnedMessage that is stored globally to modify the current topics
    if add == true {
        if pinnedMessage.Content == "This is where your reminder topics will be stored!" {
            pinnedMessage, _ = discord.ChannelMessageEdit(ReminderChannelID, pinnedMessage.ID, topic)
            dailies[topic] = false
            newJsonString, err := json.Marshal(dailies)
            if err != nil {
                log.Fatal("could not jsonify dailies map")
            }
            os.WriteFile("reminders.json", newJsonString, 0644)
            return
        }
        newContent := pinnedMessage.Content + " " + topic
        pinnedMessage, _ = discord.ChannelMessageEdit(ReminderChannelID, pinnedMessage.ID, newContent)
        dailies[topic] = false
        newJsonString, err := json.Marshal(dailies)
        if err != nil {
            log.Fatal("could not jsonify dailies map")
        }
        os.WriteFile("reminders.json", newJsonString, 0644)
        return
    }

    if pinnedMessage.Content == "This is where your reminder topics will be stored!" {
        println("Nothing to remove")
        return
    }
    // need to split the string held by pinnedMessage.Content, and remove the 
    // string that has the topic, then join it again and modify the pinned message
    currTopicsString := pinnedMessage.Content
    currTopics := strings.Split(currTopicsString, " ")
    if len(currTopics) == 1 {
        newContent := "This is where your reminder topics will be stored!"
        pinnedMessage, _ = discord.ChannelMessageEdit(ReminderChannelID, pinnedMessage.ID, newContent)
        delete(dailies, topic)
        newJsonString, err := json.Marshal(dailies)
        if err != nil {
            log.Fatal("could not jsonify dailies map")
        }
        os.WriteFile("reminders.json", newJsonString, 0644)
        return
    }

    for i, tpc := range currTopics {
        if tpc == topic {
            currTopics[i] = ""
        }
    }

    newContent := strings.Join(currTopics, " ")
    pinnedMessage, _ = discord.ChannelMessageEdit(ReminderChannelID, pinnedMessage.ID, newContent)
    delete(dailies, topic)
    newJsonString, err := json.Marshal(dailies)
    if err != nil {
        log.Fatal("could not jsonify dailies map")
    }
    os.WriteFile("reminders.json", newJsonString, 0644)
    return
}

func markDailyCompleted(topic string) bool {

    // Given a topic, update the map 'dailies' and set dailies[topic] to true,
    // then update the json file to mark it as true as well.
    if _, ok := dailies[topic]; ok {
        dailies[topic] = true
        newJsonString, err := json.Marshal(dailies)
        if err != nil {
            log.Fatal("could not jsonify dailies map")
        }
        os.WriteFile("reminders.json", newJsonString, 0644)
        return true

    } else {

        fmt.Println("Key not found in list of reminder topics")
        return false

    }
}

func getUnfinishedTopics() []*discordgo.ApplicationCommandOptionChoice {

    currUnfinished := []*discordgo.ApplicationCommandOptionChoice{}

    err := readTopicsFromJSON()
    if err != nil {
        log.Fatal("could not read from JSON in getUnfinishedTopics")
    }

    for topic, finished := range dailies {
        if finished == false {
            currUnfinished = append(currUnfinished, &discordgo.ApplicationCommandOptionChoice{Name: topic, Value: topic})
        }
    }
    return currUnfinished
}

func getAllTopics() []*discordgo.ApplicationCommandOptionChoice {

    currTopics := []*discordgo.ApplicationCommandOptionChoice{}

    err := readTopicsFromJSON()
    if err != nil {
        log.Fatal("could not read from JSON in getAllTopics")
    }

    for topic := range dailies {
        currTopics = append(currTopics, &discordgo.ApplicationCommandOptionChoice{Name: topic, Value: topic})
    }
    return currTopics

}

func resetReminders() {

    // set all values in dailies to false and then write it to the json file
    for k := range dailies {
        dailies[k] = false
    }

    newJsonString, err := json.Marshal(dailies)
    if err != nil {
        log.Fatal("could not jsonify dailies map")
    }
    os.WriteFile("reminders.json", newJsonString, 0644)

}

func updateCommandOptions(session *discordgo.Session, guildID string, commandID string, options []*discordgo.ApplicationCommandOption) error {
    command, err := session.ApplicationCommandEdit(session.State.User.ID, guildID, commandID, &discordgo.ApplicationCommand{
        Options: options,
    })

    if err != nil { 
        return err 
    }

    log.Printf("Updated command: %s\n", command.Name)
    return nil
}

// TODO: rework the storing of the message and json, right now each function is kind of doing it separately so a helper to simplify everything would be great.
func readTopicsFromJSON() (error) {

    jsonFile, err := os.Open("reminders.json")

    if err != nil {
        log.Fatal("couldn't open reminders.json ", err)
        return err
    }

    defer jsonFile.Close()

    byteValue, _ := io.ReadAll(jsonFile)

    if len(byteValue) != 0 {
        // this modifies dailies don't really need to get the returned dailies, since
        // it can just lead to inconsistent values maybe.
        json.Unmarshal(byteValue, &dailies)
        println("Found written values in the json file")
    }

    return nil

}

func updateFinishedCommand(s *discordgo.Session) error {
    // First, get all the application commands
    commands, err := s.ApplicationCommands(s.State.User.ID, "")
    if err != nil {
        return fmt.Errorf("failed to fetch commands: %w", err)
    }

    log.Printf("Fetched %d commands", len(commands))

    // Find the 'finished' command
    var finishedCmd *discordgo.ApplicationCommand
    for _, cmd := range commands {
        log.Printf("Found command: %s", cmd.Name)
        if cmd.Name == "finished" {
            finishedCmd = cmd
            break
        }
    }

    if finishedCmd == nil {
        return fmt.Errorf("'finished' command not found")
    }

    log.Printf("Updating 'finished' command with ID: %s", finishedCmd.ID)

    // Now update the command
    _, err = s.ApplicationCommandEdit(s.State.User.ID, "", finishedCmd.ID, &discordgo.ApplicationCommand{
        Name:        "finished",
        Description: "Mark a task as finished",
        Options: []*discordgo.ApplicationCommandOption{
            {
                Type:        discordgo.ApplicationCommandOptionString,
                Name:        "task",
                Description: "The task to mark as finished",
                Required:    true,
                Choices:     getUnfinishedTopics(),
            },
        },
    })

    if err != nil {
        return fmt.Errorf("failed to update 'finished' command: %w", err)
    }

    log.Println("Successfully updated 'finished' command")

    return nil
}

func sendReminders(s *discordgo.Session, introMessage string) {
    unfinishedTasks := getUnfinishedTopics()
    if len(unfinishedTasks) > 0 {
        var taskNames []string
        for _, task := range unfinishedTasks {
            taskNames = append(taskNames, task.Name)
        }
        
        message := fmt.Sprintf("@everyone: %s\n%s", introMessage, strings.Join(taskNames, "\n"))
        
        _, err := s.ChannelMessageSend(ReminderChannelID, message)
        if err != nil {
            log.Printf("Error sending reminder: %v", err)
        }
    }
}

func scheduleReset(s *discordgo.Session) {
    for {
        now := time.Now()
        est, _ := time.LoadLocation("America/New_York")
        resetTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, est)
        if now.After(resetTime) {
            resetTime = resetTime.Add(24 * time.Hour)
        }
        time.Sleep(time.Until(resetTime))
        performReset(s)
    }
}

func scheduleReminders(s *discordgo.Session, schedule ReminderSchedule) {
    for {
        now := time.Now()
        est, _ := time.LoadLocation("America/New_York")
        reminderTime := time.Date(now.Year(), now.Month(), now.Day(), schedule.Hour, schedule.Minute, 0, 0, est)
        if now.After(reminderTime) {
            reminderTime = reminderTime.Add(24 * time.Hour)
        }
        time.Sleep(time.Until(reminderTime))
        sendReminders(s, schedule.Message)
    }
}

func performReset(s *discordgo.Session) {
    resetReminders()
    _, err := s.ChannelMessageSend(ReminderChannelID, "All tasks have been reset for the new day.")
    if err != nil {
        log.Printf("Error sending reset message: %v", err)
    }
    err = updateFinishedCommand(s)
    if err != nil {
        log.Print("Error updating finished command after reset: %v", err)
    }
}

func startScheduledTasks(s *discordgo.Session) {
    for _, schedule := range reminderSchedules {
        go scheduleReminders(s, schedule)
    }
    go scheduleReset(s)
}

type ReminderSchedule struct {
    Hour   int
    Minute int
    Message string
}

var reminderSchedules = []ReminderSchedule{
    {Hour: 8, Minute: 0, Message: "Good morning! Here are your tasks for today:"},
    {Hour: 20, Minute: 0, Message: "Evening reminder! Don't forget to complete these tasks:"},
}

