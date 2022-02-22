package main

import (
	"context"
	"egg/api"
	"egg/datastore"
	"fmt"
)

var EIUID string
var DiscordUsername string

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

	backup, err := api.GetBackupFromAPI(EIUID)
	if err != nil {
		panic(err)
	}

	record, err := api.AddUserToDatabase(ctx, dStore, backup, DiscordUsername)
	if err != nil {
		panic(err)
	}

	eb, se, err := api.GetEBandSE(record)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Info for %s:\n--> EB: %s\n--> SE: %s\n", DiscordUsername, eb, se)
}
