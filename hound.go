package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/soundhound/houndify-sdk-go/houndify"
)

const (
	houndifyClientId  = "***REMOVED***"
	houndifyClientKey = "***REMOVED***"
)

// Creates a pseudo unique / random request ID.
func createRequestID() string {
	n := 10
	b := make([]byte, n)
	rand.Read(b)
	return fmt.Sprintf("%X", b)
}

func userId() string {
	userId := os.Getenv("USER")
	if len(userId) == 0 {
		userId = "hound"
	}
	return userId
}

func main() {
	flagVerbose := flag.Bool("v", false, "Verbose mode")
	flag.Parse()

	// Make log not print out time info
	log.SetFlags(0)

	houndifyClient := houndify.Client{
		ClientID:  houndifyClientId,
		ClientKey: houndifyClientKey,
		Verbose:   *flagVerbose,
	}
	houndifyClient.EnableConversationState()

	query := strings.Join(flag.Args(), " ")
	if len(query) == 0 {
		query = "what can you do?"
	}

	req := houndify.TextRequest{
		Query:             query,
		UserID:            userId(),
		RequestID:         createRequestID(),
		RequestInfoFields: make(map[string]interface{}),
	}

	serverResponse, err := houndifyClient.TextSearch(req)
	if err != nil {
		log.Fatalf("Unable to talk to houndify: %v\n%s\n", err, serverResponse)
	}
	writtenResponse, err := houndify.ParseWrittenResponse(serverResponse)
	if err != nil {
		log.Fatalf("Failed to understand houndify's response:\n%s\n", serverResponse)
	}
	fmt.Println(writtenResponse)
}
