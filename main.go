package main

import (
	"encoding/json"
	"fmt"
	"github.com/skratchdot/open-golang/open"
	"github.com/zmb3/spotify"
	"io/ioutil"
	"os"
	"os/user"
	"path"
)

var configDir = ".echoespl"

type EchoesConfig struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	AuthToken    string `json:"current_token"`
}

func getAuthenticatedClient(config EchoesConfig, forceRefresh bool) (string, spotify.Client) {
	var token string
	var auth spotify.Authenticator

	if config.AuthToken == "" || forceRefresh {
		url, channel, newAuth := StartAuthenticationFlow(config.ClientId, config.ClientSecret)
		auth = newAuth
		fmt.Println("Opening", url)

		open.Start(url)

		token = <-channel
	} else {
		auth = GetDefaultAuthenticator(config.ClientId, config.ClientSecret)
		token = config.AuthToken
	}
	oauthToken, err := auth.Exchange(token)
	if err != nil && forceRefresh == false {
		return getAuthenticatedClient(config, true)
	} else if err != nil {
		panic(err)
	}

	client := auth.NewClient(oauthToken)

	return oauthToken.RefreshToken, client
}

// accessible returns whether the given file or directory is accessible or not
func accessible(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return false
}

func getConfig() EchoesConfig {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	configPath := path.Join(user.HomeDir, configDir)

	if !accessible(configPath) {
		os.Mkdir(configPath, 0700)
	}

	filePath := path.Join(configPath, "config")

	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		panic(err)
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	var cfg EchoesConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		panic(err)
	}

	return cfg
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("The URL to an Echoes playlist must be provided")
		os.Exit(1)
	}

	url := os.Args[len(os.Args)-1]
	shows, err := GetShows(url)

	config := getConfig()

	if err != nil {
		fmt.Println("There's been a problem: %v", err)
		os.Exit(1)
	}

	for _, show := range shows {
		fmt.Println(show.Title, "|", show.Album, "|", show.Artist)
	}

	token, _ := getAuthenticatedClient(config, false)

	fmt.Println("Got", token)
}
