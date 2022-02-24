# egg
A place to fool around with the Egg, Inc. API

### Requirements
1. go version 1.16 or later
2. Your EI user ID from the game
3. A file at the root of the repo named 'config.json'

### Setup
#### Install dependencies
`go mod vendor`

#### Populate `config.json`
```
{
  "botToken": "<discord bot token>",
  "guildID": "<discord server guild id>"
}
```

### Run code start a discord bot
`go run *.go`

### Current commands
`/register` - Requires a string as input. The expected value is a user's Egg, Inc. user ID
`/removeid` - Requires a string as input. The expected value is a user's Egg, Inc. user ID

### Run tests
From the root of the repo, run `go test ./...`
