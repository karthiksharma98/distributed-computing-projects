package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	// TODO: read config files
	// TODO: join -> sendmessage to introducer to add to list, gets list in return
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

	configFile := ReadConfig()
	serviceInfo := configFile["service"]
	service := serviceInfo.(map[string]interface{})

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
			addr, _ := service["introducer_ip"].(string)
			port := service["port"]

			// Initialize new membership list
			Initialize()
			fmt.Println("Memberlist created.")

			// Add self as member
			member := NewMember(addr)
			fmt.Printf("Added introducer: ")
			fmt.Println(member)
			fmt.Printf("Membership list: ")
			fmt.Println(GetAllMembers())

			// Start listening
			if str, ok := port.(string); ok {
				Listener(str)
			}
			fmt.Println("Node has joined the group.")

		case "join":
			// Join()
			fmt.Println("Node has joined the group.")

		case "leave":
			// 	Leave()
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
