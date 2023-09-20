package main

import (
	"flag"
	"log"
	"regexp"

	"ece428/src/service"
)

func main() {
	isClientPtr := flag.Bool("client", false, "configure process as client")
	isServerPtr := flag.Bool("server", false, "configure process as server")
	expressionPtr := flag.String("expression", "", "regular expression")
	// configFilePtr := flag.String("config", "", "Location of Config File")

	flag.Parse()

	// os.Setenv("CONFIG", *configFilePtr)

	switch {
	case *isServerPtr:
		service.Server()
	case *isClientPtr:
		if *expressionPtr == "" {
			log.Fatalln("Must specify expression if using as a client")
		}
		_, err := regexp.Compile(*expressionPtr)
		if err != nil {
			log.Fatalln(err)
		}
		service.Client(*expressionPtr)
	default:
		log.Fatalln("Usage: main [-client] [-server] [--expression=<regular expression>]")
	}
}
