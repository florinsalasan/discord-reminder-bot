# discord-reminder-bot

Using streaks is a powerful motivator for a person to continue using a 
product or servicer. Unfortunately I've come up with too many streaks to 
keep track of so I am writing this. It is a discord bot in a personal server
where I can set a number of tasks I want to perform daily and be reminded of 
them periodically throughout the day.

## Data flow (in my mind)

So the private discord that the bot is added to will have 3 text channels, a 
main channel that has listeners for user instructions and where the bot will 
send reminder messages to, a reminder-topics channel that should contain a
single pinned message that has the list of items that the bot is setting
reminders for, and finally a channel that tracks which tasks have been
marked as done and gets reset every night.
### 
bot.go should control all of the logic flow, and then main is just to run it
and coordinate between the env variables and bot.go file.

## TODO:

- [ ] Everything, currently this is not a functioning project
- [x] Connect to discord api using discordgo package from bwmarrin
- [ ] Find a method of storing the list of reminders a user wants to be reminded of
- [ ] Setup the cron jobs for periodic reminders while tasks are not marked as complete
- [ ] Setup a way to mark tasks as complete for the day
- [ ] Setup a bot command to add and remove daily tasks to be reminded of
- [ ] Host it somewhere not on my computer to let it run all the time
- [ ] Instead of pulling the reminder topics from a pinned message, consider a db
