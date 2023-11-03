package main

import (
	"log"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/multiplay/go-ts3"
)

type ChainLink interface {
	Run(context *BotContext, next func())
	Name() string
}

type Chain struct {
	links []ChainLink
}

type LogLink struct{}

func (link LogLink) Run(context *BotContext, next func()) {
  update := context.update
	log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
	next()
}
func (link LogLink) Name() string {
	return "LogLink"
}

type CommandLink struct {
	commands  map[string]CommandHandler
}

func NewCommandLink() CommandLink {
	return CommandLink{
		commands: make(map[string]CommandHandler),
	}
}
func (link CommandLink) Run(context *BotContext, next func()) {
  update := context.update
	text := update.Message.Text
  telegram := context.telegram
  respond := func(text string) {
    msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
    msg.ReplyToMessageID = update.Message.MessageID
    telegram.Send(msg)
  }
	if text == "" {
		return
	}
	if text[0] != '/' {
    respond("Type /help to see available commands")
		return
	}
	command := text[1:]
  cmd, args := parseCommand(command)

	if handler, ok := link.commands[cmd]; ok {
		if handler.IsAdmin() && !context.IsAdmin() {
      respond("You are not allowed to use this command")
			return
		}
		if handler.IsRestricted() && !context.IsOnWhitelist() {
      respond("You are not allowed to use this command")
      return
		}
		handler.Run(args, respond, context)
	} else {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown command")
		msg.ReplyToMessageID = update.Message.MessageID
		telegram.Send(msg)
	}
}
func (link CommandLink) Name() string {
	return "CommandLink"
}
func (link CommandLink) AddCommand(handler CommandHandler) {
	link.commands[handler.Command()] = handler
}

type BotContext struct {
	telegram  *tgbotapi.BotAPI
	teamspeak *ts3.Client
  update *tgbotapi.Update
  config *Config
  repository *Repository
}

func (context BotContext) IsAdmin() bool {
  id := context.update.SentFrom().ID
  for _, admin_id := range context.config.Bot.AdminIds {
    if admin_id == id {
      return true
    }
  }
  return false
}

func (context BotContext) IsOnWhitelist() bool {
  is_on_whitelist, err := context.repository.IsOnWhitelist(context.update.SentFrom().ID)
  return (err == nil && is_on_whitelist) || context.IsAdmin()
}

func (context BotContext) GetUserID() int64 {
  return context.update.SentFrom().ID
}

func onMessage(context BotContext, chain *Chain) {
	if len(chain.links) == 0 {
		log.Println("Chain is empty")
		return
	}
	var counter *int = new(int)
	*counter = 0
	var next func()
	next = func() {
		if *counter >= len(chain.links) {
			return
		}
		c := *counter
		*counter++
		log.Println("Running link", chain.links[c].Name())
		chain.links[c].Run(&context, next)
	}
	next()
}
