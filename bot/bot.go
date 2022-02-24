package bot

import (
	"context"
	"egg/api"
	"egg/config"
	"egg/datastore"
	"errors"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "register",
			Description: "Register an Egg, Inc. user ID with the bot",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "id",
					Description: "Your Egg, Inc. user ID",
					Required:    true,
				},
			},
		},
		{
			Name:        "removeid",
			Description: "Remove an Egg, Inc. user ID from the bot",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "id",
					Description: "Your Egg, Inc. user ID",
					Required:    true,
				},
			},
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, store datastore.Database, ctx context.Context){
		"register": func(s *discordgo.Session, i *discordgo.InteractionCreate, store datastore.Database, ctx context.Context) {
			eggID := strings.TrimSpace(i.ApplicationCommandData().Options[0].StringValue())

			backup, err := api.GetBackupFromAPI(eggID)
			if err != nil {
				sendErrToDiscord(s, i, err)
				return
			}
			if backup.EiUserId != eggID {
				sendErrToDiscord(s, i, errors.New(fmt.Sprintf(":exclamation: '%s' isn't a recognized user ID :exclamation:", eggID)))
				return
			}

			if _, err = api.AddUserToDatabase(ctx, store, backup, i.Member.User.Username); err != nil {
				sendErrToDiscord(s, i, err)
				return
			}

			if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   1 << 6,
					Content: fmt.Sprintf(":tada: Congratulations! You've successfully registered %s with the bot :tada:", eggID),
				},
			}); err != nil {
				sendErrToDiscord(s, i, err)
			}
		},
		"removeid": func(s *discordgo.Session, i *discordgo.InteractionCreate, store datastore.Database, ctx context.Context) {
			eggID := strings.TrimSpace(i.ApplicationCommandData().Options[0].StringValue())

			if err := api.RemoveUserFromDatabase(ctx, store, eggID, i.Member.User.Username); err != nil {
				sendErrToDiscord(s, i, err)
				return
			}

			if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   1 << 6,
					Content: fmt.Sprint(":frowning2: Sad to see you go, but your account has been successfully removed from the bot :frowning2:"),
				},
			}); err != nil {
				sendErrToDiscord(s, i, err)
			}
		},
	}
)

// Start initializes the Discord bot by adding handlers and registering commands
func Start(ctx context.Context, store datastore.Database) ([]*discordgo.ApplicationCommand, *discordgo.Session) {
	s, err := discordgo.New(fmt.Sprintf("Bot %s", config.Config.Token))
	if err != nil {
		panic(err)
	}

	u, err := s.User("@me")
	if err != nil {
		panic(err)
	}

	logrus.Infof("--> logged in as %v#%v", u.Username, u.Discriminator)

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i, store, ctx)
		}
	})

	if err = s.Open(); err != nil {
		panic(err)
	}

	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, command := range commands {
		cmd, cmdCreateErr := s.ApplicationCommandCreate(s.State.User.ID, config.Config.GuildID, command)
		if cmdCreateErr != nil {
			panic(cmdCreateErr)
		}

		registeredCommands[i] = cmd
	}

	logrus.Info("--> bot is running")

	return registeredCommands, s
}

func sendErrToDiscord(s *discordgo.Session, i *discordgo.InteractionCreate, input error) {
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:   1 << 6,
			Content: input.Error(),
		},
	}); err != nil {
		logrus.Fatal(err)
	}
}
