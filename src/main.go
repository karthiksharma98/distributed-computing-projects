package main

import (
        "os"
)

func main() {
        // TODO: read config files
        // TODO: wait for input to query operations on node?
        // TODO: join -> sendmessage to introducer to add to list, gets list in return
        // TODO: start listening after recieving membershiplist, announce to random member or smth
        
        // Test send/recv UDP packet
        // Start a listener somewhere with ./main listen <port>
        // Send a text message: ./main send <ip>:<port> <message>
        arg := os.Args[1]

        switch arg {
        case "send":
                SendMessage(os.Args[2], os.Args[3])
        case "listen":
                Listener(os.Args[2])
        }

}
