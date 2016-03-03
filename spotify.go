package main

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	redirect_url      = "http://localhost:5533/auth_callback"
	redirect_port     = ":5533"
	redirect_url_part = "/auth_callback"
)

func GetDefaultAuthenticator(clientID string, secretKey string) spotify.Authenticator {
	auth := spotify.NewAuthenticator(redirect_url, spotify.ScopePlaylistReadPrivate, spotify.ScopePlaylistModifyPrivate)
	auth.SetAuthInfo(clientID, secretKey)
	return auth
}

type AuthenticationResponse struct {
	TokenResponseChannel chan oauth2.Token
	TokenResponseError   chan error
	Authenticator        spotify.Authenticator
	ClientRedirectUri    string
}

// StartAuthentication begins an authentication request to Spotify. It returns
// an authentication URL and a channel. The channel will publish a the token when
// the http server receives the callback.
func StartAuthenticationFlow(clientID string, secretKey string) (AuthenticationResponse, error) {
	auth := GetDefaultAuthenticator(clientID, secretKey)
	state := generateState()
	authUrl := auth.AuthURL(state)

	replyChannel := make(chan oauth2.Token, 1)
	errorChannel := make(chan error, 1)

	http.HandleFunc(redirect_url_part, func(w http.ResponseWriter, r *http.Request) {
		transferredState, stateOk := r.URL.Query()["state"]
		code, codeOk := r.URL.Query()["code"]

		if !stateOk || !codeOk || transferredState[0] != state {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Not quite what we're looking for"))
		}

		token, err := auth.Exchange(code[0])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Sorry, something's gone wrong in the gubbins!"))
			errorChannel <- err
		} else {
			w.Write([]byte("All done. Close your browser now."))
			replyChannel <- *token
		}
	})

	go http.ListenAndServe(redirect_port, nil)

	return AuthenticationResponse{replyChannel, errorChannel, auth, authUrl}, nil
}

func generateState() string {
	now := time.Now().UTC().UnixNano()
	rand.Seed(now + int64(os.Getpid()))
	compiledString := []byte(fmt.Sprintf(string(now), os.Getpid(), rand.Int()))
	hash := md5.New().Sum(compiledString)
	return base64.StdEncoding.EncodeToString(hash)
}

// BuildPlaylist creates a Spotify playlist and returns a playlist identifier. authedClient is
// an authenticated Spotify client, and songs is a list of songs scraped from the Echoes website
func BuildPlaylist(authedClient *spotify.Client, songs []EchoesSong) (string, error) {
	terms := echoesSongsToSearchStrings(songs)
	results := make([]chan *spotify.SearchResult, len(songs))
	errors := make([]chan error, len(songs))

	// Dispatch each request in parallel
	for _, term := range terms {
		term := term
		replyChannel := make(chan *spotify.SearchResult)
		errorChannel := make(chan error)

		results = append(results, replyChannel)
		errors = append(errors, errorChannel)

		go func(t string, reply chan *spotify.SearchResult, errorReply chan error) {
			result, err := authedClient.Search(t, spotify.SearchTypeTrack)
			if err != nil {
				errorReply <- err
			} else {
				reply <- result
			}
		}(term, replyChannel, errorChannel)
	}

	// Collate the responses
	for i := range results {
		replyChannel := results[i]
		errorChannel := errors[i]

		select {
		case result := <-replyChannel:
			tracks := result.Tracks.Total

			fmt.Println("Got a result with", tracks, "tracks")
		case err := <-errorChannel:
			fmt.Println("Got an error:", err)
		}
	}

	fmt.Println(len(terms), len(results), len(errors))
	return "", nil
}

func echoesSongsToSearchStrings(songs []EchoesSong) []string {
	terms := make([]string, len(songs))
	for _, v := range songs {
		terms = append(terms, echoesSongToSearchString(v))
	}
	return terms
}

func echoesSongToSearchString(song EchoesSong) string {
	searchStrings := prependSearchType("title", song.Title)
	searchStrings = append(searchStrings, prependSearchType("album", song.Album)...)
	searchStrings = append(searchStrings, prependSearchType("artist", song.Artist)...)

	return strings.Join(searchStrings, " ")
}

func prependSearchType(searchType, title string) []string {
	terms := make([]string, 0)
	for _, s := range strings.Split(title, " ") {
		if strings.TrimSpace(s) != "" {
			term := fmt.Sprintf("%v:%v", searchType, s)
			terms = append(terms, term)
		}
	}
	return terms
}
