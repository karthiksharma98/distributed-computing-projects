package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

var (
	// Configuration stores all info in config.json
	Configuration Config
	process       *Member
)

func printOptions() {
	if process == nil {
		fmt.Println("Welcome! Don't be a loner and join the group by saying \"join introducer\" or \"join\".")
	} else {
		if !enabledHeart {
			fmt.Println("Start heartbeating with \"start\".")
		}

		fmt.Print("Interact with the group using any of the following: leave, kill, ")
		fmt.Println("status, get logs {-n}, grep {all}, stop, switch (gossip/alltoall), or chat")
	}
}

func main() {
	// Set up loggers and configs
	InitLog()
	Configuration = ReadConfig()
	Configuration.Print()

	for {
		printOptions()
		// wait for input to query operations on node
		consoleReader := bufio.NewReader(os.Stdin)
		fmt.Print("> ")
		input, _ := consoleReader.ReadString('\n')
		input = strings.ToLower(strings.TrimSuffix(input, "\n"))
		inputFields := strings.Fields(input) // Split string into os.Args like array

		if len(inputFields) == 0 {
			Info.Println("invalid command")
			continue
		}

		switch inputFields[0] {
		case "join":
			if process != nil {
				Warn.Println("You have already joined!")
				continue
			}

			if len(inputFields) == 2 && inputFields[1] == "introducer" {
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
			if process == nil {
				Warn.Println("You need to join in order to leave!")
				continue
			}

			Info.Println("Node has left the group.")
			process = nil

		case "kill":
			// simulate a failure?
			Warn.Println("Killing process. Bye bye.")
			os.Exit(1)

		case "status":
			// TODO
			if process == nil {
				Warn.Println("You need to join in order to get status!")
				continue
			}

			process.PrintMembershipList(os.Stdout)
			Info.Println("[imagine some status here].")

		case "get":
			if len(inputFields) >= 2 && inputFields[1] == "logs" {
				if len(inputFields) == 4 && inputFields[2] == "-n" {
					num, err := strconv.Atoi(inputFields[3])
					if err != nil {
						fmt.Println("Please provide a valid number of lines")
						continue
					}

					printLogs(num)

				} else {
					printLogs(0)
				}
			}

		case "grep":
			if process == nil {
				Warn.Println("You need to join in order to get status!")
				continue
			}

			if len(inputFields) >= 2 {
				if len(inputFields) >= 3 && inputFields[1] == "all" {
					process.Grep(strings.Join(inputFields[2:], " "), false)
				} else {
					process.Grep(strings.Join(inputFields[1:], " "), true)
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
				Warn.Println("You need to join group before you start.")
				continue
			}

			go process.Tick()

		case "stop":
			if !enabledHeart {
				Warn.Println("No process running to stop.")
				continue
			}

			disableHeart <- true

		case "switch":
			if len(inputFields) >= 2 && inputFields[1] == "gossip" {
				if isGossip {
					Warn.Println("You are already running Gossip")
					continue
				}

				SetHeartbeating(true)
				process.SendAll(SwitchMsg, []byte{1})
			} else if len(inputFields) >= 2 && inputFields[1] == "alltoall" {
				if !isGossip {
					Warn.Println("You are already running All to All")
					continue
				}

				SetHeartbeating(false)
				process.SendAll(SwitchMsg, []byte{0})
			}

		default:
			Info.Println("invalid command")
		}
	}

}
