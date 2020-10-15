package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	// Configuration stores all info in config.json
	Configuration Config
	process       *Member
	client        *rpc.Client
)

func printOptions() {
	if process == nil {
		fmt.Println("Welcome! Don't be a loner and join the group by saying \"join introducer\" or \"join\".")
	} else {
		fmt.Print("Interact with the group using any of the following: leave, kill, ")
		fmt.Println("status, get logs {-n}, grep {all}, stop, switch (gossip/alltoall), or chat")
	}
}

func main() {
	// Set up loggers and configs
	InitLog()
	InitMonitor()
	Configuration = ReadConfig()
	Configuration.Print()

	rpcInitialized := false

	for {
		printOptions()
		// wait for input to query operations on node
		consoleReader := bufio.NewReader(os.Stdin)
		fmt.Print("> ")
		input, _ := consoleReader.ReadString('\n')
		input = strings.ToLower(strings.TrimSuffix(input, "\n"))
		inputFields := strings.Fields(input) // Split string into os.Args like array

		if len(inputFields) == 0 {
			fmt.Println("invalid command")
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

				if Configuration.Service.detectorType == "alltoall" {
					isGossip = false
				}

				process.membershipList[0] = NewMembershipListEntry(0, net.ParseIP(Configuration.Service.introducerIP))
				go process.Listen(fmt.Sprint(Configuration.Service.port))
				Info.Println("You are now the introducer.")

				// register RPC server
				if process != nil {
					if rpcInitialized == false {
						err := rpc.Register(process)
						if err != nil {
							fmt.Println("Format isn't correct. ", err)
						}
						rpc.HandleHTTP()
						rpcListener, e := net.Listen("tcp", ":9092")
						if e != nil {
							fmt.Println("error in starting listener")
						}

						fmt.Printf("Serving RPC server on port %d\n", 9092)
						// Start accepting incoming HTTP connections
						go http.Serve(rpcListener, nil)
						rpcInitialized = true
					}
				}

			} else {
				// Temporarily, the memberID is 0, will be set to correct value when introducer adds it to group
				process = NewMember(false)
				if Configuration.Service.detectorType == "alltoall" {
					isGossip = false
				}

				go process.Listen(fmt.Sprint(Configuration.Service.port))
				time.Sleep(100 * time.Millisecond) // Sleep a tiny bit so listener can start
				process.joinRequest()
				// Wait for response
				select {
				case _ = <-joinAck:
					fmt.Println("Node has joined the group.")
				case <-time.After(2 * time.Second):
					fmt.Println("Timeout join. Please retry again.")
					listener.Close()
					process = nil
				}

				if rpcInitialized == false {
					var err error
					client, err = rpc.DialHTTP("tcp", "172.22.156.42:9092")
					if err != nil {
						fmt.Println("Connection error: ", err)
					}
					rpcInitialized = true
				}
			}

			// start gossip
			if process == nil {
				Warn.Println("You need to join group before you start.")
				continue
			}

			go process.Tick()

		case "leave":
			if process == nil {
				Warn.Println("You need to join in order to leave!")
				continue
			}
			process.leave()
			listener.Close()
			Info.Println("Node has left the group.")
			process = nil

		case "kill":
			Warn.Println("Killing process. Bye bye.")
			os.Exit(1)

		case "status":
			if process == nil {
				Warn.Println("You need to join in order to get status!")
				continue
			}

			process.PrintMembershipList(os.Stdout)

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
					// sleep for a tiny bit of time so that you get all results before restarting loop
					time.Sleep(250 * time.Millisecond)
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

		case "stop":
			process.StopTick()

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

		// Monitoring
		case "metrics":
			if memMetrics == nil {
				InitMonitor()
			}
			memMetrics.PrintMonitor()

		case "whoami":
			if process == nil {
				Warn.Println("You need to join group before you are assigned an ID.")
				continue
			}

			fmt.Println("You are member " + fmt.Sprint(process.memberID))

		case "sim":
			if len(inputFields) >= 2 && inputFields[1] == "failtest" {
				process.SendAll(TestMsg, []byte{})
			}

		case "putfile":
			if len(inputFields) >= 3 {
				req := SdfsRequest{LocalFName: inputFields[1], RemoteFName: inputFields[2], Type: PutReq}
				var res SdfsResponse

				err := client.Call("Member.HandlePutRequest", req, &res)

				if err != nil {
					fmt.Println("putfile failed", err)
				} else {
					// TODO: copy local file to store as well

					// TODO: Initiate upload to each alive node
					for _, val := range res.IPList {
						fmt.Println(val)
						// call upload

						// TODO: post each successful upload to an IP, call Member.AddIPToFileMap so master can add IP to list of IPs containing file
						// client.Call("Member.AddIPToFileMap", UploadAck{RemoteFName: <file_name>, IPAddr: <process_ip>}, <pointer to res>)
					}
				}
			}

		case "getfile":
			if len(inputFields) >= 3 {
				req := SdfsRequest{LocalFName: inputFields[2], RemoteFName: inputFields[1], Type: GetReq}
				var res SdfsResponse

				err := client.Call("Member.HandleGetRequest", req, &res)

				if err != nil {
					fmt.Println(err)
				} else {
					fmt.Println(res.IPList)
					// TODO: Choose one of the received IPs and initiate download

					// TODO: after successful download, call Member.AddIPToFileMap so master can add to list of IPs containing file
					// client.Call("Member.AddIPToFileMap", UploadAck{RemoteFName: <file_name>, IPAddr: <process_ip>}, <pointer to res>)
				}

			}

		case "deletefile":
			if len(inputFields) >= 2 {
				req := SdfsRequest{LocalFName: "", RemoteFName: inputFields[1], Type: DelReq}
				var res SdfsResponse

				err := client.Call("Member.HandleDeleteRequest", req, &res)

				if err != nil {
					fmt.Println(err)
				} else {
					fmt.Println("Deleted successfully:", req.RemoteFName)
				}
			}

		default:
			fmt.Println("invalid command")
		}
	}

}
