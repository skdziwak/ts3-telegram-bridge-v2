# ts3-telegram-bridge-v2

This bot serves as a bridge between a Teamspeak server and a Telegram chat. It allows users to get notifications about Teamspeak server events and get Teamspeak info from Telegram.

## Features

- Notifications for client connect and disconnects from Teamspeak to Telegram
- Telegram command interface for managing bot features
- User subscription to specific Teamspeak client notifications
- Whitelist management for restricting bot command usage
- Quote management system with Teamspeak integration

## Requirements

- Go (programming language)
- Access to a Teamspeak server with query access
- A Telegram bot token (obtained by talking to @BotFather on Telegram)
- Access to a MongoDB instance for storing data

## Configuration

The application requires a `config.yaml` file to store different configuration settings like Telegram token, Teamspeak host parameters, admin IDs, and MongoDB URI.

Example of `config.yaml` structured content:

```yaml
bot:
  telegram_token: "your-telegram-token"
  teamspeak_host: "ts01.teamspeak.com"
  teamspeak_user: "user"
  teamspeak_password: "password"
  teamspeak_port: 9892
  teamspeak_query_port: 21004
  admin_ids:
    - 1234
  mongodb_uri: "mongodb://root:example@localhost:27017/"
```

## Usage

1. Configure the `config.yaml` as described above.
2. Run the bot with command `go run .` in the directory where the files are located.

### Admin Commands

Admin users can manage whitelist entries, update and delete quotes, and set the Teamspeak channel where the quotes are posted.

```
/whitelist add <id> <alias> - Adds a new entry to the bot's whitelist
/whitelist remove <id> - Removes an entry from the bot's whitelist
/whitelist list - Lists all entries in the bot's whitelist
/updatequotes - Updates the quotes on the Teamspeak server
/deletequote <uuid> - Deletes a quote by UUID
/setquotechannel <channel name> - Sets the Teamspeak channel for posting quotes
```

### General Commands

All users on the whitelist can use the following commands:

```
/help - Prints the help message with available commands
/me - Prints your Telegram ID
/list [id] [all] - List online Teamspeak users; can show IDs and all users
/subscribe <Teamspeak id> - Subscribe to notifications for a specific Teamspeak user
/subscribed - List all subscribed Teamspeak users
/unsubscribe <Teamspeak id> - Unsubscribe from a specific Teamspeak user
/addquote <author> <content> - Adds a new quote
/listquotes [id] - Lists all quotes, with optional UUID display
/exportquotes - Exports all quotes to a text file and sends it in the chat
```

### Event Notifications

The bot listens to the Teamspeak server and sends notifications to Telegram chats when users connect or disconnect from Teamspeak.

## Internal Mechanisms

The bot is implemented in Go and uses packages such as:

- `github.com/go-telegram-bot-api/telegram-bot-api/v5` for Telegram Bot API interactions
- `github.com/multiplay/go-ts3` for Teamspeak Server Query API interactions

Bot functionality is organized using Command Handlers, Chain of Responsibility for command processing, and concurrent notification listening.

## Teamspeak References and Notifications

The application interacts with a Teamspeak server using Server Query commands to retrieve user lists, manage notifications, and update a channel with quotes.

## Notes

Database interactions for managing whitelists, subscribers, and quotes happen through a MongoDB instance. Be sure to have the database running and accessible using the provided MongoDB URI in the config file.

Ensure the bot token and Teamspeak credentials provided are correct and that the Teamspeak server has query access enabled.
