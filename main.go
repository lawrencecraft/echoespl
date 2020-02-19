package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"

	"github.com/skratchdot/open-golang/open"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

var configDir = ".echoespl"

// EchoesConfig holds configuration
type EchoesConfig struct {
	ClientID     string        `json:"client_id"`
	ClientSecret string        `json:"client_secret"`
	AuthToken    *oauth2.Token `json:"current_token"`
}

func getAuthenticatedClient(config EchoesConfig, forceRefresh bool) (*spotify.Client, error) {
	auth := GetDefaultAuthenticator(config.ClientID, config.ClientSecret)
	if config.AuthToken == nil || forceRefresh {
		authenticationResponse, err := StartAuthenticationFlow(config.ClientID, config.ClientSecret)
		if err != nil {
			return nil, err
		}

		// Redirect user to the authentication URL
		url := authenticationResponse.ClientRedirectURI
		fmt.Println("Please visit", url, "if your browser does not automatically start")
		open.Start(url)

		select {
		case tokenError := <-authenticationResponse.TokenResponseError:
			return nil, tokenError
		case token := <-authenticationResponse.TokenResponseChannel:
			client := auth.NewClient(&token)
			config.AuthToken = &token

			err = saveConfig(config)
			if err != nil {
				// Don't end, just write the error
				fmt.Println("There's a problem saving the configuration file")
			}

			return &client, nil
		}
	} else {
		client := auth.NewClient(config.AuthToken)
		return &client, nil
	}
}

// accessible returns whether the given file or directory is accessible or not
func accessible(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return false
}

func getConfig() (EchoesConfig, error) {
	filePath := getConfigPath()
	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		return EchoesConfig{}, err
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return EchoesConfig{}, err
	}

	var cfg EchoesConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return EchoesConfig{}, err
	}

	return cfg, nil
}

func saveConfig(config EchoesConfig) error {
	configPath := getConfigPath()

	bytes, err := json.Marshal(config)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(configPath, bytes, 0600)
	if err != nil {
		return err
	}
	return nil
}

func getConfigPath() string {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	configPath := path.Join(user.HomeDir, configDir)

	if !accessible(configPath) {
		os.Mkdir(configPath, 0700)
	}

	filePath := path.Join(configPath, "config")
	return filePath
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("The URL to an Echoes playlist must be provided")
		os.Exit(1)
	}

	str := flag.String("p", "", "Default playlist name")
	refresh := flag.Bool("r", false, "-r forces echoespl to refresh the OAuth token")
	countryCode := flag.String("c", "GB", "-c determines the country code. Defaults to GB")
	flag.Parse()

	url := os.Args[len(os.Args)-1]
	songs, err := GetShows(url)

	if err != nil {
		fmt.Println("There was a problem retrieving shows:", err)
		os.Exit(1)
	}

	config, err := getConfig()

	if err != nil {
		fmt.Println("There's been a problem:", err)
		os.Exit(1)
	}

	for _, song := range songs {
		fmt.Println(song.Title, "|", song.Album, "|", song.Artist)
	}
	fmt.Println("That's", len(songs), "songs")

	client, err := getAuthenticatedClient(config, *refresh)
	if err != nil {
		fmt.Println("Problem authenticating:", err)
		os.Exit(1)
	}

	playlist, err := BuildPlaylist(client, *str, songs, *countryCode)

	if err != nil {
		fmt.Println("Error building playlist:", err)
		os.Exit(1)
	}

	fmt.Println("Playlist with ID", playlist, "successfully created. Happy listening!")
}
