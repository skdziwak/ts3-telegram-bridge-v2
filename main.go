package main

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/multiplay/go-ts3"
)

func main() {
  config := load_config()
  host := fmt.Sprintf("%s:%d", config.Bot.TeamspeakHost, config.Bot.TeamspeakQueryPort)
	teamspeak, err := ts3.NewClient(host)
	if err != nil {
		log.Panic(err)
	}
	defer teamspeak.Close()
  if err := teamspeak.Login(config.Bot.TeamspeakUser, config.Bot.TeamspeakPassword); err != nil {
		log.Panic(err)
	}
  if err := teamspeak.UsePort(config.Bot.TeamspeakPort); err != nil {
    log.Panic(err)
  } else {
    log.Println("Connected to Teamspeak server on port", config.Bot.TeamspeakPort)
  }
	if v, err := teamspeak.Version(); err != nil {
		log.Panic(err)
	} else {
		log.Println("Connected to Teamspeak server version", v.Version)
	}

	telegram, err := tgbotapi.NewBotAPI(config.Bot.TelegramToken)
	if err != nil {
		log.Panic(err)
	}
	telegram.Debug = true
	log.Printf("Authorized on account %s", telegram.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	telegram_updates := telegram.GetUpdatesChan(u)

  repository, err := CreateRepository(&config)
  if err != nil {
    log.Panic(err)
  }

	command_link := NewCommandLink()
	RegisterCommands(&command_link)
	chain := Chain{
		links: []ChainLink{
			&LogLink{},
			&command_link,
		},
	}

  notifications_context := NotificationsContext{
    teamspeak: teamspeak,
    repository: repository,
    telegram: telegram,
  }
  receive_notifications(&notifications_context)
	for update := range telegram_updates {
		if update.Message != nil {
      context := BotContext{
        telegram: telegram,
        teamspeak: teamspeak,
        update: &update,
        config: &config,
        repository: repository,
      }
			onMessage(context, &chain)
		}
	}
}
