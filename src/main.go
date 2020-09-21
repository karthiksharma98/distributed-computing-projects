package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

var (
	// Configuration stores all info in config.json
	Configuration Config
	process       *Member
)

func main() {
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
		args := strings.Fields(input) // Split string into os.Args like array

		if len(args) == 0 {
			Info.Println("invalid command")
			continue
		}

		switch args[0] {
		case "join":
			if len(args) == 2 && args[1] == "introducer" {
				process = NewMember(true)
				process.membershipList[0] = NewMembershipListEntry(0, net.ParseIP(Configuration.Service.introducerIP))
				go process.Listen(fmt.Sprint(Configuration.Service.port))
				Info.Println("You are now the introducer.")
			} else {
				// Temporarily, the memberID is 0, will be set to correct value when introducer adds it to group
				process = NewMember(false)
				process.joinRequest()
				go process.Listen(fmt.Sprint(Configuration.Service.port))
				Info.Println("Node has joined the group.")
			}
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
			process.PrintMembershipList(os.Stdout)
			Info.Println("[imagine some status here].")
		case "get":
			if len(args) >= 2 && args[1] == "logs" {
				if len(args) == 3 && args[2] == "-a" {
					Info.Println("[imagine some logs logs here].")
				} else {
					Info.Println("[imagine some logs here].")
				}
			}
		case "chat":
			// FOR DEBUGGING PURPOSES
			// DEBUG addresses
			addresses := []string{"172.22.156.42:9000", "172.22.158.42:9000", "172.22.94.42:9000", "172.22.156.43:9000"}
			Info.Println("Joined chat")
			for {
				consoleReader := bufio.NewReader(os.Stdin)
				fmt.Print("> ")
				input, _ := consoleReader.ReadString('\n')
				SendBroadcast(addresses, 2, []byte(input))
			}
		case "start":
			if process == nil {
				Warn.Println("You are not in a group.")
			}
			go process.Tick()
		case "stop":
			if enabledHeart == true {
				disableHeart <- true
			}
		case "switch":
			if args[1] == "gossip" {
				SetHeartbeating(true)
			} else if args[1] == "alltoall" {
				SetHeartbeating(false)
			}
		default:
			Info.Println("invalid command")
		}
	}

}
