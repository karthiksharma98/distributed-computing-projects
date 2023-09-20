package service

import (
	"log"
	"mp0-go-2019/src/config"
	"mp0-go-2019/src/finder"
	"net/http"
	"net/rpc"
	"path/filepath"
)

// Response stores a list of "hits" returned back to the RPC caller
type Response struct {
	Matches []finder.MatchRes
}

// Request stores the contents of a RPC request
type Request struct {
	Expression string
}

// Logly is a type that implements the Retrieve() "method"
type Logly struct{}

// Retrieve is a RPC style function which accepts a regular expression and returns a list of "hits" in it's local logfile.
func (l *Logly) Retrieve(request *Request, response *Response) error {
	log.Println("Retrieve:", request.Expression)
	//Find local log file
	gMatches, _ := filepath.Glob("./" + "vm[0-9]*.*log")
	fileName := gMatches[0]
	_, fileName = filepath.Split(fileName)
	matches, err := finder.Finder(request.Expression, fileName)
	if err != nil {
		return err
	}

	response.Matches = matches
	return nil
}

// Server is a long running function that accepts new RPC requests and fufills them
func Server() {
	//Get Port
	port, err := config.Port()
	if err != nil {
		log.Fatalln("Failed to get port:", err)
	}

	log.Println("Listening on port", port)

	// Setup logly service
	logly := new(Logly)
	rpc.Register(logly)
	rpc.HandleHTTP()

	//Start Listening on defined port
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalln("ListenAndServe:", err)
	}
}
