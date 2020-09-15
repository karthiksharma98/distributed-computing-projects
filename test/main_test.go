package main

import (
        "os"
)

func main() {
        // TODO: read config files
        // TODO: wait for input to query operations on node?
        // TODO: join -> sendmessage to introducer to add to list, gets list in return
        
        // Test read packet
        // Test write packet
        arg := os.Args[1]

        switch arg {
        case "send":
                SendMessage(os.Args[2], os.Args[3])
        case "recieve":
                ListenAll(os.Args[2])
        }

}
