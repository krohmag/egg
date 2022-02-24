package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/sirupsen/logrus"
)

var (
	Config Bot
)

// Bot is the values required to run the bot
type Bot struct {
	Token   string `json:"botToken"`
	GuildID string `json:"guildID"`
}

// LoadConfigFromFile loads configuration from a file into memory
func LoadConfigFromFile(filename string) error {
	logrus.Info(fmt.Sprintf("--> reading config file: %s ...", filename))
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(file, &Config); err != nil {
		return err
	}

	return nil
}
