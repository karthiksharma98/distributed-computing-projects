package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

// Configuration stores all info in config.json
var Configuration map[string]interface{}

// ServiceInfo stores info in config.json's "service" key
var ServiceInfo map[string]interface{}

func main() {
	// TODO: wait for input to query operations on node?
	// TODO: start listening after recieving membershiplist, announce to random member or smth

	// Test send/recv UDP packet
	// Start a listener somewhere with ./main listen <port>
	// Send a text message: ./main send <ip>:<port> <message>

	if len(os.Args) > 2 {
		arg := os.Args[1]

		switch arg {
		case "send":
			SendMessage(os.Args[2], os.Args[3])
		case "listen":
			Listener(os.Args[2])
		}
	}

	Configuration = ReadConfig()
	ServiceInfo = Configuration["service"].(map[string]interface{})

	// wait for input to query operations on node
	fmt.Println("Listening for input.")
	fmt.Println("Options: join introducer, join, leave, kill, status, get logs {-a}.")
	for {
		consoleReader := bufio.NewReader(os.Stdin)
		fmt.Print("> ")
		input, _ := consoleReader.ReadString('\n')

		input = strings.ToLower(strings.TrimSuffix(input, "\n"))

		switch input {
		case "join introducer":
			process := Member{0, true, make(map[uint8]membershipListEntry)}
			process.membershipList[0] = membershipListEntry{0, net.ParseIP(ServiceInfo["introducer_ip"].(string)), 0, time.Now(), Alive}
			process.Listen(fmt.Sprint(ServiceInfo["port"]))

		case "join":
			// Temporarily, the memberID is 0, will be set to correct value when introducer adds it to group
			process := Member{0, false, make(map[uint8]membershipListEntry)}
			process.joinRequest()
			process.Listen(fmt.Sprint(ServiceInfo["port"]))
			fmt.Println("Node has joined the group.")

		case "leave":
			// 	Leave()
			// TODO: Call Member.leave() here
			fmt.Println("Node has left the group.")

		case "kill":
			// simulate a failure?
			fmt.Println("Killing process. Bye bye.")
			os.Exit(1)

		case "status":
			// TODO
			fmt.Println("[imagine some status here].")

		case "get logs -a":
			// TODO
			fmt.Println("[imagine some logs here].")

		case "get logs":
			// TODO
			fmt.Println("[imagine some logs here].")
		}
	}

}
