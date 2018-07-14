package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/casimir/xdg-go"
	"github.com/soundhound/houndify-sdk-go/houndify"
	"github.com/spf13/viper"
)

const (
	appName           = "hound-cli"
	houndifyClientId  = "***REMOVED***"
	houndifyClientKey = "***REMOVED***"
)

// Creates a pseudo unique / random string.
func randString() string {
	n := 10
	b := make([]byte, n)
	rand.Read(b)
	return fmt.Sprintf("%X", b)
}

// Creates a valid user ID.
func userId() string {
	userId := os.Getenv("USER")
	if len(userId) == 0 {
		userId = "hound-" + randString()
	}
	return userId
}

func readConfig() *viper.Viper {
	// Ensure config file exists to pacify Viper
	xdgApp := xdg.App{Name: appName}
	configFile := xdgApp.ConfigPath(appName + ".yaml")

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		configFileDir := filepath.Dir(configFile)
		if err := os.MkdirAll(configFileDir, 0755); err != nil {
			panic(err)
		}
		if f, err := os.Create(configFile); err != nil {
			panic(err)
		} else {
			defer f.Close()
		}
	}

	// Load configuration
	v := viper.New()
	v.SetConfigName(appName)
	v.AddConfigPath(xdgApp.ConfigPath("")) // path to look for the config file in
	v.SetConfigType("yaml")

	v.SetDefault("User", userId())

	err := v.ReadInConfig() // Find and read the config file
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	return v
}

func main() {
	flagVerbose := flag.Bool("v", false, "Verbose mode")
	flag.Parse()

	log.SetFlags(0)

	// Configuration
	config := readConfig()

	// Houndify
	houndifyClient := houndify.Client{
		ClientID:  houndifyClientId,
		ClientKey: houndifyClientKey,
		Verbose:   *flagVerbose,
	}
	houndifyClient.EnableConversationState()
	//TODO: houndifyClient.SetConversationState()

	query := strings.Join(flag.Args(), " ")
	if len(query) == 0 {
		query = "what can you do?"
	}

	req := houndify.TextRequest{
		Query:             query,
		UserID:            config.GetString("User"),
		RequestID:         randString(),
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

	//TODO: config.Set("State", houndifyClient.GetConversationState())
	config.WriteConfig()
}
