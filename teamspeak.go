package main

import (
	"errors"
	"log"
	"strings"

	"github.com/multiplay/go-ts3"
)

type TeamspeakUser struct {
	TsId     string
	Nickname string
}

func getTeamspeakUsers(client *ts3.Client) ([]TeamspeakUser, error) {
	list, err := client.Exec("clientlist")
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, nil
	}
	var data string = list[0]
	var lines []string = strings.Split(data, "|")
	var users []TeamspeakUser = make([]TeamspeakUser, len(lines))
	for i, line := range lines {
		var fields []string = strings.Split(line, " ")
		for _, field := range fields {
			var kv []string = strings.Split(field, "=")
			if len(kv) != 2 {
				continue
			}
			if kv[0] == "client_database_id" {
				users[i].TsId = kv[1]
			}
			if kv[0] == "client_nickname" {
				nickname := kv[1]
				nickname = strings.Replace(nickname, "\\s", " ", -1)
				users[i].Nickname = nickname
			}
		}
	}
	return users, nil
}

func getAllTeamspeakUsers(client *ts3.Client) ([]TeamspeakUser, error) {
	list, err := client.Exec("clientdblist")
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, nil
	}
	var data string = list[0]
	var lines []string = strings.Split(data, "|")
	var users []TeamspeakUser = make([]TeamspeakUser, len(lines))
	for i, line := range lines {
		var fields []string = strings.Split(line, " ")
		for _, field := range fields {
			var kv []string = strings.Split(field, "=")
			if len(kv) != 2 {
				continue
			}
			if kv[0] == "cldbid" {
				users[i].TsId = kv[1]
			}
			if kv[0] == "client_nickname" {
				nickname := kv[1]
				nickname = strings.Replace(nickname, "\\s", " ", -1)
				users[i].Nickname = nickname
			}
		}
	}
	return users, nil
}

func setChannelDescription(channel_name string, description string, client *ts3.Client) error {
	list, err := client.Exec("channellist")
	if err != nil {
		return err
	}
	log.Println("Setting channel description for", channel_name, "to", description)
	if len(list) == 0 {
		return errors.New("No channels found")
	}
	var data string = list[0]
	var lines []string = strings.Split(data, "|")
	for _, line := range lines {
		var channel struct {
			Cid  string
			Name string
		}
		var fields []string = strings.Split(line, " ")
		for _, field := range fields {
			var kv []string = strings.Split(field, "=")
			if len(kv) != 2 {
				continue
			}
			if kv[0] == "cid" {
				channel.Cid = kv[1]
			}
			if kv[0] == "channel_name" {
				channel.Name = kv[1]
			}
		}
		if channel.Name == channel_name {
			cmd := ts3.NewCmd("channeledit").WithArgs(
				ts3.NewArg("cid", channel.Cid),
				ts3.NewArg("channel_description", description),
			)
			log.Println("Executing", cmd)
			_, err := client.ExecCmd(cmd)
			if err != nil {
				return err
			}
			return nil
		}
	}

	return errors.New("Channel not found")
}

func updateTeamspeakQuotes(repository *Repository, teamspeak *ts3.Client) error {
	channel_name, ok := repository.GetQuotesChannel()
	if !ok {
		return errors.New("No quotes channel configured")
	}
	quotes, err := repository.GetAllQuotes()
	if err != nil {
		return err
	}
	description, err := createTeamspeakQuotesString(quotes)
	if err != nil {
		return err
	}
	err = setChannelDescription(channel_name, description, teamspeak)
	return nil
}

func createTeamspeakQuotesString(quotes []Quote) (string, error) {
	m := make(map[string][]string)
	var description string = ""
	for _, quote := range quotes {
		if existingQuotes, ok := m[quote.Author]; ok {
			m[quote.Author] = append(existingQuotes, quote.Content)
		} else {
			m[quote.Author] = []string{quote.Content}
		}
	}
	for author := range m {
		description += author + "\n"
		for _, quote := range m[author] {
			description += "-" + quote + "\n"
		}
	}
	return description, nil
}
