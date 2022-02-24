package main

import (
	"context"
	"egg/bot"
	"egg/config"
	"egg/datastore"
	"os"
	"os/signal"

	"github.com/sirupsen/logrus"
)

func main() {
	ctx := context.Background()
	db, err := datastore.ConnectDatabase("sqlite-file", true)
	if err != nil {
		panic(err)
	}

	if err = db.AutoMigrate(datastore.User{}); err != nil {
		panic(err)
	}

	dStore := datastore.Database{DB: db}

	if err = config.LoadConfigFromFile("./config.json"); err != nil {
		panic(err)
	}

	commands, session := bot.Start(ctx, dStore)
	defer func() {
		_ = session.Close()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	logrus.Info("--> removing bot commands from server ...")
	for _, command := range commands {
		if err = session.ApplicationCommandDelete(session.State.User.ID, config.Config.GuildID, command.ID); err != nil {
			panic(err)
		}
	}

	logrus.Info("--> gracefully shutting down ...")
}
