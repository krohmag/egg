# egg
A place to fool around with the Egg, Inc. API

### Requirements
1. go version 1.16 or later
2. Your EI user ID from the game

### Setup
#### Install dependencies
`go mod vendor`

### Run code to get EB and SE count
`go run -ldflags "-X 'main.EIUID="<your EI user ID for the game>"'" *.go`

### Expected output format
`EB: <human readable number with unit>; SE: <human readable number with unit>`

#### Example
`EB: 124.779s; SE: 4.000Q`
