package main

import (
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/casimir/xdg-go"
	"github.com/soundhound/houndify-sdk-go/houndify"
	"github.com/spf13/viper"
)

const (
	appName             = "houndify-cli"
	houndifyClientId    = "***REMOVED***"
	houndifyClientKey   = "***REMOVED***"
	ipstackApiAccessKey = "***REMOVED***"
	ipstackApiUrl       = "http://api.ipstack.com"
	ipifyApiUrl         = "https://api.ipify.org?format=json"
	unitsMetric         = "metric"
	unitsImperial       = "imperial"
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
	v.SetDefault("Units", unitsMetric)

	err := v.ReadInConfig() // Find and read the config file
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	return v
}

func unitsToHoundify(units string) string {
	switch units {
	case "imperial":
		return "US"
	default:
		return "METRIC"
	}
}

func httpGet(url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return body
}

type ExternalIp struct {
	IP string
}

type Location struct {
	Latitude     float32
	Longitude    float32
	Country_Name string
}

func location() Location {
	// External IP first
	var externalIP ExternalIp
	json.Unmarshal(httpGet(ipifyApiUrl), &externalIP)

	// Geolocaiton next
	body := httpGet(ipstackApiUrl + "/" + externalIP.IP + "?access_key=" + ipstackApiAccessKey)
	var location Location
	json.Unmarshal(body, &location)
	return location
}

func main() {
	log.SetFlags(0)

	unitSystems := map[string]bool{unitsMetric: true, unitsImperial: true}

	// Command line arguments
	flagVerbose := flag.Bool("v", false, "Verbose mode")
	flagUnits := flag.String("units", "", "Unit system, '"+unitsMetric+"' or '"+unitsImperial+"'")
	flag.Parse()

	// Configuration
	config := readConfig()

	if *flagUnits != "" {
		if _, ok := unitSystems[*flagUnits]; ok != true {
			panic(*flagUnits + " not supported")
		}
		config.Set("Units", *flagUnits)
	}

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

	location := location()

	requestInfoFields := make(map[string]interface{})
	requestInfoFields["UnitPreference"] = unitsToHoundify(config.GetString("Units"))
	requestInfoFields["Latitude"] = location.Latitude
	requestInfoFields["Longitude"] = location.Longitude
	requestInfoFields["Country"] = location.Country_Name

	req := houndify.TextRequest{
		Query:             query,
		UserID:            config.GetString("User"),
		RequestID:         randString(),
		RequestInfoFields: requestInfoFields,
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
