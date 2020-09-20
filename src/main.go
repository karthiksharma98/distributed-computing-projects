package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

// Configuration stores all info in config.json
var (
	Configuration Config
	process       *Member
)

func main() {
	// TODO: wait for input to query operations on node?
	// TODO: start listening after recieving membershiplist, announce to random member or smth
	// Set up loggers and configs
	Log(os.Stdout, os.Stdout, os.Stderr)
	Configuration = ReadConfig()
	Configuration.Print()
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
			process = NewMember(true)
			process.membershipList[0] = NewMembershipListEntry(0, net.ParseIP(Configuration.Service.introducerIP))
			go process.Listen(fmt.Sprint(Configuration.Service.port))
			Info.Println("You are now the introducer.")
		case "join":
			// Temporarily, the memberID is 0, will be set to correct value when introducer adds it to group
			process = NewMember(false)
			process.joinRequest()
			go process.Listen(fmt.Sprint(Configuration.Service.port))
			Info.Println("Node has joined the group.")
		case "leave":
			// 	Leave()
			// TODO: Call Member.leave() here
			Info.Println("Node has left the group.")
		case "kill":
			// simulate a failure?
			Warn.Println("Killing process. Bye bye.")
			os.Exit(1)

		case "status":
			// TODO
			Info.Println("[imagine some status here].")

		case "get logs -a":
			// TODO
			Info.Println("[imagine some logs here].")

		case "get logs":
			// TODO
			Info.Println("[imagine some logs here].")
		// FOR DEBUGGING PURPOSES
		case "chat -a":
			// DEBUG addresses
			addresses := []string{"172.22.156.42:9000", "172.22.158.42:9000", "172.22.94.42:9000", "172.22.156.43:9000"}
			Info.Println("Joined chat")
			for {
				consoleReader := bufio.NewReader(os.Stdin)
				fmt.Print("> ")
				input, _ := consoleReader.ReadString('\n')
				SendBroadcast(addresses, 2, []byte(input))
			}
		// Gossip/heartbeating
		case "start":
			if process == nil {
				Warn.Println("You are not in a group.")
			}
			go process.Tick()
		case "stop":
			if enabledHeart == true {
				disableHeart <- true
			}
		case "switch gossip":
			SetHeartbeating(true)
		case "switch alltoall":
			SetHeartbeating(false)
		}
	}

}
