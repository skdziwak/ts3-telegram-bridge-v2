package main

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type Config struct {
	Bot struct {
		TelegramToken      string  `yaml:"telegram_token"`
		TeamspeakHost      string  `yaml:"teamspeak_host"`
		TeamspeakPort      int     `yaml:"teamspeak_port"`
		TeamspeakQueryPort int     `yaml:"teamspeak_query_port"`
		TeamspeakUser      string  `yaml:"teamspeak_user"`
		TeamspeakPassword  string  `yaml:"teamspeak_password"`
		AdminIds           []int64 `yaml:"admin_ids"`
		MongodbUri         string  `yaml:"mongodb_uri"`
	} `yaml:"bot"`
}

func load_config() Config {
	config := Config{}

	data, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatal(err)
	}

	if config.Bot.TelegramToken == "" {
		log.Fatal("Telegram token not set")
	}
	if config.Bot.TeamspeakHost == "" {
		log.Fatal("Teamspeak host not set")
	}
	if config.Bot.TeamspeakPort == 0 {
		log.Fatal("Teamspeak port not set")
	}
	if config.Bot.TeamspeakQueryPort == 0 {
		log.Fatal("Teamspeak query port not set")
	}
	if config.Bot.TeamspeakUser == "" {
		log.Fatal("Teamspeak user not set")
	}
	if config.Bot.TeamspeakPassword == "" {
		log.Fatal("Teamspeak password not set")
	}
	if len(config.Bot.AdminIds) == 0 {
		log.Fatal("Admin IDs not set")
	}
	if config.Bot.MongodbUri == "" {
		log.Fatal("MongoDB URI not set")
	}

	return config
}
