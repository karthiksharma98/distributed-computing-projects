package main

/*
import _ "net/http/pprof"
import "net/http"
*/

import (
	"bufio"
	"fmt"
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
	sdfs          *SdfsNode
	client        *rpc.Client
)

func printOptions() {
	if process == nil {
		fmt.Println("Welcome! Don't be a loner and join the group by saying one of the following:\n" +
			"-	join introducer\n" +
			"-	join")
	} else {
		fmt.Println("\nGroup Interaction Options:")
		fmt.Println("-	leave | kill | stop | switch [gossip/alltoall]")

		fmt.Println("\nQuery Group Information:")
		fmt.Println("-	status | whoami | print logs {-n} | grep {all} [query]")

		fmt.Println("\nResource Monitoring:")
		fmt.Println("-	metrics | sim failtest")

		fmt.Println("\nSDFS Commands:")
		fmt.Println("-	put [local file] [sdfsfile]")
		fmt.Println("-	get [sdfs file] [local file]")
		fmt.Println("-	delete [sdfs file]")
		fmt.Println("-	ls")
		fmt.Println("-	store")
	}
}

func main() {
	// Set up loggers and configs
	InitLog()
	InitMonitor()
	Configuration = ReadConfig()
	Configuration.Print()
	InitSdfsDirectory()
	printOptions()
	for {
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
			if process != nil && sdfs != nil {
				Warn.Println("You have already joined!")
				continue
			}


			if len(inputFields) == 2 && inputFields[1] == "introducer" {
                                process = InitMembership(true)
                                sdfs = InitSdfs(process, true)
				// initialize file transfer server
				go InitializeServer(fmt.Sprint(Configuration.Service.filePort))


			} else {
				// Temporarily, the memberID is 0, will be set to correct value when introducer adds it to group
                                process = InitMembership(false)
                                sdfs = InitSdfs(process, false)
				go InitializeServer(fmt.Sprint(Configuration.Service.filePort))
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

		case "print":
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

		case "stop":
			process.StopTick()

		case "switch":
			if len(inputFields) >= 2 && inputFields[1] == "gossip" {
				SetHeartbeating(true)
				process.SendAll(SwitchMsg, []byte{1})
			} else if len(inputFields) >= 2 && inputFields[1] == "alltoall" {
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

		// SDFS
		case "put":
			if len(inputFields) >= 3 && process != nil {
                                /*
                                go func() {
                                        Info.Println(http.ListenAndServe("localhost:6060", nil))
                                }()
                                */
				if client == nil || sdfs == nil {
					Warn.Println("Client not initialized.")
					continue
				}
				sessionId := sdfs.RpcLock(int32(sdfs.Member.memberID), inputFields[2], SdfsLock)
				req := SdfsRequest{LocalFName: inputFields[1], RemoteFName: inputFields[2], Type: PutReq}

				numAlive := process.GetNumAlive()
				numSuccessful := 0
				ipsAttempted := make(map[string]bool)
				// attempt to get as many replications needed, until you've attempted all the IPs
				for numSuccessful < int(Configuration.Settings.replicationFactor) &&
					len(ipsAttempted) <= numAlive {
					var res SdfsResponse
					var err error
					if len(ipsAttempted) == 0 {
						err = client.Call("SdfsNode.HandlePutRequest", req, &res)
					} else {
						err = client.Call("SdfsNode.GetRandomNodes", req, &res)
					}

					if err != nil {
						fmt.Println("Put failed", err)
						break
					} else {
						// attempt upload each file
						for _, ipAddr := range res.IPList {
							if _, exists := ipsAttempted[ipAddr.String()]; !exists {
								ipsAttempted[ipAddr.String()] = true
								err := Upload(ipAddr.String(), fmt.Sprint(Configuration.Service.filePort), req.LocalFName, req.RemoteFName)

								if err != nil {
									fmt.Println("error in upload process.", err)
								} else {
									numSuccessful += 1
									// succesfull upload -> add to master's file map
									mapReq := SdfsRequest{LocalFName: ipAddr.String(), RemoteFName: inputFields[2], Type: AddReq}
									var mapRes SdfsResponse
									mapErr := client.Call("SdfsNode.ModifyMasterFileMap", mapReq, &mapRes)
									if mapErr != nil {
										fmt.Println(mapErr)
									}
								}
							}
						}
					}

					// update alive nodes in case there's not enough anymore
					numAlive = process.GetNumAlive()
				}
				sessionId = sdfs.RpcUnlock(sessionId, inputFields[2], SdfsLock)

				fmt.Println("Finished put.")
			}

		case "get":
			if len(inputFields) >= 3 {
				if client == nil || sdfs == nil {
					Warn.Println("Client not initialized.")
					continue
				}
				sessionId := sdfs.RpcLock(int32(sdfs.Member.memberID), inputFields[1], SdfsRLock)
				req := SdfsRequest{LocalFName: inputFields[2], RemoteFName: inputFields[1], Type: GetReq}
				var res SdfsResponse

				err := client.Call("SdfsNode.HandleGetRequest", req, &res)
				if err != nil {
					fmt.Println(err)
				} else {
					for _, ipAddr := range res.IPList {
						err := Download(ipAddr.String(), fmt.Sprint(Configuration.Service.filePort), req.RemoteFName, req.LocalFName)

						if err != nil {
							fmt.Println("error in download process.", err)
						} else {
							// successful download
							break
						}
					}
				}
				sessionId = sdfs.RpcUnlock(sessionId, inputFields[1], SdfsRLock)
				fmt.Println("Finished get.")
			}

		case "delete":
			if len(inputFields) >= 2 {
				if client == nil || sdfs == nil {
					Warn.Println("Client not initialized.")
					continue
				}
				sessionId := sdfs.RpcLock(int32(sdfs.Member.memberID), inputFields[1], SdfsLock)
                                sdfs.RpcDelete(inputFields[1])
				sessionId = sdfs.RpcUnlock(sessionId, inputFields[1], SdfsLock)
			}

		case "ls":
			if len(inputFields) >= 2 {
				if client == nil {
					Warn.Println("Client not initialized.")
					continue
				}
                                sdfs.RpcListIPs(inputFields[1])
			}

		case "store":
			if sdfs == nil {
				fmt.Println("SDFS not initialized.")
				continue
			}
			sdfs.Store()

		case "upload":
			if len(inputFields) == 4 {
				Upload(fmt.Sprint(inputFields[1]),
					fmt.Sprint(Configuration.Service.filePort),
					inputFields[2],
					inputFields[3])
			}

		case "download":
			if len(inputFields) == 4 {
				err := Download(fmt.Sprint(inputFields[1]),
					fmt.Sprint(Configuration.Service.filePort),
					inputFields[2],
					inputFields[3])
				if err != nil {
					fmt.Println(err)
				}
			}

		case "master":
			if sdfs != nil {
				fmt.Println(sdfs.MasterId)
			}
		// Debug lock
		case "lock":
			if sdfs == nil {
				return
			}
			if len(inputFields) == 3 {
				if inputFields[1] == "get" {
					// Acquire read
					sessionId := sdfs.RpcLock(int32(sdfs.Member.memberID), inputFields[2], SdfsRLock)
					fmt.Println("Lock (get) acquired! Test get for 10 seconds!")

					// Timeout
					time.Sleep(10 * time.Second)

					// Release read
					sessionId = sdfs.RpcUnlock(sessionId, inputFields[2], SdfsRLock)
					fmt.Println("Lock read released!")
				} else if inputFields[1] == "put" {
					// Acquire write
					sessionId := sdfs.RpcLock(int32(sdfs.Member.memberID), inputFields[2], SdfsLock)
					fmt.Println("Lock (write) acquired! Test put for 10 seconds!")

					// Timeout
					time.Sleep(10 * time.Second)

					// Release write
					sessionId = sdfs.RpcUnlock(sessionId, inputFields[2], SdfsLock)
				}
			}

		case "help":
			printOptions()
		default:
			fmt.Println("invalid command")
			printOptions()
		}
	}

}
