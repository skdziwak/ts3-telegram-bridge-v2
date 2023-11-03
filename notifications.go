package main

import (
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/multiplay/go-ts3"
)

type NotificationsContext struct {
  teamspeak *ts3.Client
  repository *Repository
  telegram *tgbotapi.BotAPI
}

func receive_notifications(notifications_context *NotificationsContext) {
  teamspeak := notifications_context.teamspeak
	go func() {
    users, err := getTeamspeakUsers(teamspeak)
    if err != nil {
      log.Println(err)
      return
    }
		notifications := teamspeak.Notifications()
		log.Println("Listening for Teamspeak notifications")
		teamspeak.Server.Register("server")
		for notification := range notifications {
      if notification.Type == "clientleftview" || notification.Type == "cliententerview" {
        log.Println("Received Teamspeak notification:", notification.Type)
        new_users, err := getTeamspeakUsers(teamspeak)
        if err != nil {
          log.Println(err)
          continue
        }
        added, removed := find_differences(users, new_users)
        users = new_users
        for _, user := range added {
          log.Println("Client connected")
          send_message_to_subscribers(user.TsId, "Client {name} connected", notifications_context)
        }
        for _, user := range removed {
          log.Println("Client disconnected")
          send_message_to_subscribers(user.TsId, "Client {name} disconnected", notifications_context)
        }
      } else {
        continue
      }
		}
	}()
}

func find_differences(old []TeamspeakUser, current []TeamspeakUser) (added []TeamspeakUser, removed []TeamspeakUser) {
    oldMap := make(map[string]TeamspeakUser)
    currentMap := make(map[string]TeamspeakUser)

    for _, user := range old {
        oldMap[user.TsId] = user
    }
    for _, user := range current {
        currentMap[user.TsId] = user
    }

    for tsId, user := range currentMap {
        if _, found := oldMap[tsId]; !found {
            added = append(added, user)
        }
    }

    for tsId, user := range oldMap {
        if _, found := currentMap[tsId]; !found {
            removed = append(removed, user)
        }
    }

    return added, removed
}

func send_message_to_subscribers(ts_id string, message string, notifications_context *NotificationsContext) {
  repository := notifications_context.repository
  telegram := notifications_context.telegram
  subscribers, err := repository.GetSubscribers(ts_id)
  if err != nil {
    log.Println("Error getting subscribers:", err)
    return
  }
  name := subscribers.Name
  message = strings.Replace(message, "{name}", name, -1)

  for _, subscriber := range subscribers.TelegramSubscribers {
    telegram.Send(tgbotapi.NewMessage(subscriber, message))
  }
}
