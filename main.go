package main

import (
	"egg/api"
	"fmt"
)

var EIUID string

func main() {
	eb, se, err := api.GetEBandSE(EIUID)
	if err != nil {
		panic(err)
	}

	fmt.Printf("EB: %s; SE: %s", eb, se)
}
