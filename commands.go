package main

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

type CommandHandler interface {
	Command() string
	Description() string
	IsAdmin() bool
	IsRestricted() bool
	Run(args []string, respond func(string), context *BotContext)
}

func RegisterCommands(link *CommandLink) {
	log.Println("Registering commands")
	link.AddCommand(HelpCommand{&link.commands})
	link.AddCommand(MeCommand{})
	link.AddCommand(ListCommand{})
	link.AddCommand(WhitelistCommand{})
	link.AddCommand(SubscribeCommand{})
	link.AddCommand(SubscribedCommand{})
	link.AddCommand(UnsubscribeCommand{})
	link.AddCommand(AddQuoteCommand{})
	link.AddCommand(UpdateQuotesCommand{})
	link.AddCommand(SetQuoteChannelCommand{})
	link.AddCommand(ListQuotesCommand{})
	link.AddCommand(DeleteQuoteCommand{})
	link.AddCommand(ExportQuotesCommand{})
}

type HelpCommand struct {
	commands *map[string]CommandHandler
}

func (cmd HelpCommand) Command() string {
	return "help"
}
func (cmd HelpCommand) Description() string {
	return "Prints this help message"
}
func (cmd HelpCommand) IsAdmin() bool {
	return false
}
func (cmd HelpCommand) IsRestricted() bool {
	return false
}
func (cmd HelpCommand) Run(args []string, respond func(string), context *BotContext) {
	var text string = "Available commands:\n"
	for _, handler := range *cmd.commands {
		if handler.IsAdmin() && !context.IsAdmin() {
			continue
		} else if handler.IsRestricted() && !context.IsOnWhitelist() {
			continue
		}
		text += "/" + handler.Command() + " - " + handler.Description() + "\n"
	}
	respond(text)
}

type MeCommand struct{}

func (cmd MeCommand) Command() string {
	return "me"
}
func (cmd MeCommand) Description() string {
	return "Prints your Telegram ID"
}
func (cmd MeCommand) IsAdmin() bool {
	return false
}
func (cmd MeCommand) IsRestricted() bool {
	return false
}
func (cmd MeCommand) Run(args []string, respond func(string), context *BotContext) {
	var id int64 = context.update.SentFrom().ID
	var id_str string = fmt.Sprintf("%d", id)
	respond("Your Telegram ID is " + id_str)
}

type ListCommand struct{}

func (cmd ListCommand) Command() string {
	return "list"
}
func (cmd ListCommand) Description() string {
	return "List all online users. Can be modified with arguments: /list id, /list all, /list id all"
}
func (cmd ListCommand) IsAdmin() bool {
	return false
}
func (cmd ListCommand) IsRestricted() bool {
	return true
}
func (cmd ListCommand) Run(args []string, respond func(string), context *BotContext) {
	showIds := false
	showAll := false
	for _, arg := range args {
		if arg == "id" {
			showIds = true
		} else if arg == "all" {
			showAll = true
		}
	}
	teamspeak := context.teamspeak
	var users []TeamspeakUser
	var err error
	if showAll {
		users, err = getAllTeamspeakUsers(teamspeak)
	} else {
		users, err = getTeamspeakUsers(teamspeak)
	}
	if err != nil {
		log.Println(err)
		respond("Error getting Teamspeak users")
		return
	}

	if len(users) == 0 {
		if showAll {
			respond("No users found")
		} else {
			respond("No users online")
		}
		return
	}

	var text string
	if showAll {
		text = "All users:\n"
	} else {
		text = "Online users:\n"
	}
	for _, user := range users {
		if showIds {
			text += " - " + user.Nickname + " (id: " + user.TsId + ")\n"
		} else {
			text += " - " + user.Nickname + "\n"
		}
	}
	respond(text)
}

type WhitelistCommand struct{}

func (cmd WhitelistCommand) Command() string {
	return "whitelist"
}
func (cmd WhitelistCommand) Description() string {
	return "Manage the whitelist. Usage: /whitelist add <id> <alias>, /whitelist remove <id>, /whitelist list"
}
func (cmd WhitelistCommand) IsAdmin() bool {
	return true
}
func (cmd WhitelistCommand) IsRestricted() bool {
	return false
}
func (cmd WhitelistCommand) Run(args []string, respond func(string), context *BotContext) {
	if len(args) == 0 {
		respond("Usage: /whitelist add <id> <alias>, /whitelist remove <id>, /whitelist list")
		return
	}
	var subcommand string = args[0]
	if subcommand == "add" {
		if len(args) != 3 {
			respond("Usage: /whitelist add <id> <alias>")
			return
		}
		var id_str string = args[1]
		var alias string = args[2]
		var id int64
		if _, err := fmt.Sscanf(id_str, "%d", &id); err != nil {
			respond("Invalid Telegram ID")
			return
		}
		if err := context.repository.AddWhiteListEntry(id, alias); err != nil {
			respond("An error occured")
			log.Println(err)
			return
		}
		respond("Added " + alias + " to the whitelist")
	} else if subcommand == "remove" {
		if len(args) != 2 {
			respond("Usage: /whitelist remove <id>")
			return
		}
		var id_str string = args[1]
		id, err := strconv.ParseInt(id_str, 10, 64)
		if err != nil {
			respond("Invalid Telegram ID")
			return
		}
		if err := context.repository.RemoveWhiteListEntry(id); err != nil {
			respond("An error occured")
			log.Println(err)
			return
		}
		respond("Removed " + id_str + " from the whitelist")
	} else if subcommand == "list" {
		if len(args) != 1 {
			respond("Usage: /whitelist list")
			return
		}
		var text string = "Whitelisted users:\n"
		entries, err := context.repository.GetWhiteList()
		if err != nil {
			respond("An error occured")
			log.Println(err)
			return
		}
		for _, entry := range entries {
			text += fmt.Sprintf("%s - %s\n", entry.Alias, entry.Id)
		}
		respond(text)
	}
}

type SubscribeCommand struct{}

func (cmd SubscribeCommand) Command() string {
	return "subscribe"
}
func (cmd SubscribeCommand) Description() string {
	return "usage: /subscribe <Teamspeak id>"
}
func (cmd SubscribeCommand) IsAdmin() bool {
	return false
}
func (cmd SubscribeCommand) IsRestricted() bool {
	return true
}
func (cmd SubscribeCommand) Run(args []string, respond func(string), context *BotContext) {
	if len(args) != 1 {
		respond("usage: /subscribe <Teamspeak id>")
		return
	}
	var identifier string = args[0]
	users, err := getAllTeamspeakUsers(context.teamspeak)
	if err != nil {
		log.Println(err)
		respond("Error getting Teamspeak users")
		return
	}
	for _, user := range users {
		if user.TsId == identifier {
			err := context.repository.AddSubscriber(context.update.SentFrom().ID, user.TsId, user.Nickname)
			if err != nil {
				log.Println(err)
				respond("Error adding subscriber")
				return
			}
			respond("Added subscriber")
			return
		}
	}
}

type SubscribedCommand struct{}

func (cmd SubscribedCommand) Command() string {
	return "subscribed"
}
func (cmd SubscribedCommand) Description() string {
	return "List all subscribed Teamspeak users"
}
func (cmd SubscribedCommand) IsAdmin() bool {
	return false
}
func (cmd SubscribedCommand) IsRestricted() bool {
	return true
}
func (cmd SubscribedCommand) Run(args []string, respond func(string), context *BotContext) {
	var text string = "Subscribed users:\n"
	entries, err := context.repository.GetSubscribedTeamspeaks(context.update.SentFrom().ID)
	if err != nil {
		respond("An error occured")
		log.Println(err)
		return
	}
	for _, entry := range entries {
		text += fmt.Sprintf("%s - %s\n", entry.Name, entry.Id)
	}
	respond(text)
}

type UnsubscribeCommand struct{}

func (cmd UnsubscribeCommand) Command() string {
	return "unsubscribe"
}
func (cmd UnsubscribeCommand) Description() string {
	return "usage: /unsubscribe <Teamspeak id>"
}
func (cmd UnsubscribeCommand) IsAdmin() bool {
	return false
}
func (cmd UnsubscribeCommand) IsRestricted() bool {
	return true
}
func (cmd UnsubscribeCommand) Run(args []string, respond func(string), context *BotContext) {
	if len(args) != 1 {
		respond("usage: /unsubscribe <Teamspeak id>")
		return
	}
	var id string = args[0]
	err := context.repository.RemoveSubscriber(context.update.SentFrom().ID, id)
	if err != nil {
		log.Println(err)
		respond("Error removing subscriber")
		return
	}
	respond("Removed subscriber")
}

type AddQuoteCommand struct{}

func (cmd AddQuoteCommand) Command() string {
	return "addquote"
}
func (cmd AddQuoteCommand) Description() string {
	return "Adds a new quote. Usage: /addquote <author> <context>"
}
func (cmd AddQuoteCommand) IsAdmin() bool {
	return false
}
func (cmd AddQuoteCommand) IsRestricted() bool {
	return true
}
func (cmd AddQuoteCommand) Run(args []string, respond func(string), context *BotContext) {
	if len(args) != 2 {
		respond("Usage: /addquote <author> <content>")
		return
	}
	author := args[0]
	content := args[1]
  content = strings.ReplaceAll(content, "\\n", "\n")
	uuid := uuid.New().String()
	quote := Quote{UUID: uuid, Author: author, Content: content, CreatedBy: context.GetUserID()}
	err := context.repository.AddQuote(quote)
	if err != nil {
		log.Println(err)
		respond("Error adding quote")
		return
	}
	respond("Quote added with ID: " + uuid)
  err = updateTeamspeakQuotes(context.repository, context.teamspeak)
  if err != nil {
    log.Println(err)
    return
  }
  respond("Updated Teamspeak quotes")
}

type UpdateQuotesCommand struct{}

func (cmd UpdateQuotesCommand) Command() string {
  return "updatequotes"
}
func (cmd UpdateQuotesCommand) Description() string {
  return "Updates the quotes on the Teamspeak server"
}
func (cmd UpdateQuotesCommand) IsAdmin() bool {
  return true
}
func (cmd UpdateQuotesCommand) IsRestricted() bool {
  return false
}
func (cmd UpdateQuotesCommand) Run(args []string, respond func(string), context *BotContext) {
	if len(args) != 0 {
		respond("Usage: /updatequotes")
		return
	}
  err := updateTeamspeakQuotes(context.repository, context.teamspeak)
  if err != nil {
    log.Println(err)
    return
  }
  respond("Updated Teamspeak quotes")
}

type ListQuotesCommand struct{}

func (cmd ListQuotesCommand) Command() string {
	return "listquotes"
}
func (cmd ListQuotesCommand) Description() string {
	return "Lists all quotes. Add 'id' to also show UUIDs: /listquotes id"
}
func (cmd ListQuotesCommand) IsAdmin() bool {
	return false
}
func (cmd ListQuotesCommand) IsRestricted() bool {
	return true
}
func (cmd ListQuotesCommand) Run(args []string, respond func(string), context *BotContext) {
	showUUIDs := len(args) > 0 && args[0] == "id"
	quotes, err := context.repository.GetAllQuotes()
	if err != nil {
		log.Println(err)
		respond("Error retrieving quotes")
		return
	}
	if len(quotes) == 0 {
		respond("No quotes found")
		return
	}
	var responseText string
	for _, quote := range quotes {
		if showUUIDs {
			responseText += fmt.Sprintf("ID: %s - \"%s\" by %s\n", quote.UUID, quote.Content, quote.Author) // Changed from Context to Content
		} else {
			responseText += fmt.Sprintf("\"%s\" by %s\n", quote.Content, quote.Author) // Changed from Context to Content
		}
	}
	respond(responseText)
}

type DeleteQuoteCommand struct{}

func (cmd DeleteQuoteCommand) Command() string {
	return "deletequote"
}
func (cmd DeleteQuoteCommand) Description() string {
	return "Deletes a quote by UUID. Usage: /deletequote <uuid>"
}
func (cmd DeleteQuoteCommand) IsAdmin() bool {
	return true
}
func (cmd DeleteQuoteCommand) IsRestricted() bool {
	return false
}
func (cmd DeleteQuoteCommand) Run(args []string, respond func(string), context *BotContext) {
	if len(args) != 1 {
		respond("Usage: /deletequote <uuid>")
		return
	}
	uuid := args[0]
	err := context.repository.DeleteQuote(uuid)
	if err != nil {
		log.Println(err)
		respond("Error deleting quote")
		return
	}
	respond("Quote deleted")
  err = updateTeamspeakQuotes(context.repository, context.teamspeak)
  if err != nil {
    log.Println(err)
    return
  }
  respond("Updated Teamspeak quotes")
}

type SetQuoteChannelCommand struct{}

func (cmd SetQuoteChannelCommand) Command() string {
  return "setquotechannel"
}
func (cmd SetQuoteChannelCommand) Description() string {
  return "Sets the channel where quotes are posted. Usage: /setquotechannel <channel name>"
}
func (cmd SetQuoteChannelCommand) IsAdmin() bool {
  return true
}
func (cmd SetQuoteChannelCommand) IsRestricted() bool {
  return false
}
func (cmd SetQuoteChannelCommand) Run(args []string, respond func(string), context *BotContext) {
  if len(args) != 1 {
    respond("Usage: /setquotechannel <channel name>")
    return
  }
  channelName := args[0]
  err := context.repository.SetQuotesChannel(channelName)
  if err != nil {
    log.Println(err)
    respond("Error setting quotes channel")
    return
  }
  respond("Quotes channel set to " + channelName)
}

type ExportQuotesCommand struct{}

func (cmd ExportQuotesCommand) Command() string {
    return "exportquotes"
}

func (cmd ExportQuotesCommand) Description() string {
    return "Exports all quotes to a text file and sends it."
}

func (cmd ExportQuotesCommand) IsAdmin() bool {
    return false
}

func (cmd ExportQuotesCommand) IsRestricted() bool {
    return true
}

func (cmd ExportQuotesCommand) Run(args []string, respond func(string), context *BotContext) {
    quotes, err := context.repository.GetAllQuotes()
    if err != nil {
        log.Println(err)
        respond("Error retrieving quotes")
        return
    }
    if len(quotes) == 0 {
        respond("No quotes found")
        return
    }

    var buffer bytes.Buffer
    for _, quote := range quotes {
        line := fmt.Sprintf("%s - \"%s\"\n", quote.Author, quote.Content)
        _, err := buffer.WriteString(line)
        if err != nil {
            log.Println(err)
            respond("Error writing quotes to buffer")
            return
        }
    }

    chatID := context.update.Message.Chat.ID
    reader := bytes.NewReader(buffer.Bytes())
    file := tgbotapi.FileReader{Name: "quotes.txt", Reader: reader }
    msg := tgbotapi.NewDocument(chatID, file)

    _, err = context.telegram.Send(msg)
    if err != nil {
        log.Println(err)
        respond("Failed to send the quotes file")
        return
    }

    respond("Quotes exported and sent successfully")
}
