# egg
A place to fool around with the Egg, Inc. API

### Requirements
1. go version 1.16 or later
2. Your EI user ID from the game

### Setup
#### Install dependencies
`go mod vendor`

### Run code to get EB and SE count
`go run -ldflags "-X 'main.EIUID="<your EI user ID for the game>"' -X 'main.DiscordUsername="<your discord username>"'" *.go`

### Expected output format
```
Info for <discord name>:
--> EB: <human readable number with unit>
--> SE: <human readable number with unit>
```

#### Example
```
Info for krohmag:
--> EB: 124.779s
--> SE: 4.000Q
```
