package main

import (
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
	"math/rand"
	"net/http"
	"os"
	"strconv"
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
	// This is not secure and should not be relied upon for more than basic validation
	now := time.Now().UTC().UnixNano()

	rand.Seed(now)
	nowString := strconv.Itoa(rand.Int())
	nextFromNow := md5.New().Sum([]byte(nowString))

	rand.Seed(int64(os.Getpid()))
	pidString := strconv.Itoa(rand.Int())
	nextFromPid := md5.New().Sum([]byte(pidString))

	finalArray := make([]byte, len(nextFromNow), len(nextFromNow))

	for i := range finalArray {
		finalArray[i] = nextFromNow[i] ^ nextFromPid[i]
	}

	return base64.StdEncoding.EncodeToString(finalArray)
}

// BuildPlaylist creates a Spotify playlist and returns a playlist identifier. authedClient is
// an authenticated Spotify client, and songs is a list of songs scraped from the Echoes website
func BuildPlaylist(authedClient *spotify.Client, songs []EchoesSong, market string) (string, error) {
	terms := echoesSongsToSearchStrings(songs)
	fmt.Println("Buildling playlist of", len(terms), "terms")
	results := make([]chan *spotify.SearchResult, 0, len(songs))
	errorResult := make([]chan error, 0, len(songs))

	tracks := make([]spotify.FullTrack, 0)

	// Dispatch each request in parallel
	for _, term := range terms {
		term := term
		replyChannel := make(chan *spotify.SearchResult)
		errorChannel := make(chan error)

		results = append(results, replyChannel)
		errorResult = append(errorResult, errorChannel)

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
		errorChannel := errorResult[i]
		song := songs[i]

		select {
		case result := <-replyChannel:
			if result.Tracks == nil {
				fmt.Println("Possible problem with track", song.Title,
					"by", song.Artist, "or the Spotify API. Please contact lawrencecraft on github")
			} else {
				for _, track := range result.Tracks.Tracks {
					if isValidForMarket(track, market) {
						tracks = append(tracks, track)
						break
					}
				}
			}
		case err := <-errorChannel:
			fmt.Println("Got an error:", err)
		}
	}

	if len(tracks) > 0 {
		return CreatePlaylist(authedClient, tracks)
	}
	return "", errors.New("No tracks found")
}

// CreatePlaylist creates a new playlist given a set of tracks with a default name
func CreatePlaylist(authedClient *spotify.Client, tracks []spotify.FullTrack) (string, error) {
	playlistName := generatePlaylistName(time.Now())

	user, err := authedClient.CurrentUser()
	if err != nil {
		return "", err
	}

	playlist, err := authedClient.CreatePlaylistForUser(user.User.ID, playlistName, false)
	if err != nil {
		return "", err
	}

	trackIds := []spotify.ID{}
	for _, track := range tracks {
		trackIds = append(trackIds, track.ID)
	}

	_, err = authedClient.AddTracksToPlaylist(user.User.ID, playlist.ID, trackIds...)
	return playlistName, err
}

func generatePlaylistName(t time.Time) string {
	// NOTE: Change before 2030
	return fmt.Sprintf("Echoespl Playlist %v", t.Unix())
}

func isValidForMarket(track spotify.FullTrack, targetMarket string) bool {
	for _, market := range track.AvailableMarkets {
		if market == targetMarket {
			return true
		}
	}

	return false
}

func echoesSongsToSearchStrings(songs []EchoesSong) []string {
	terms := make([]string, 0, len(songs))
	for _, v := range songs {
		terms = append(terms, echoesSongToSearchString(v))
	}
	return terms
}

func echoesSongToSearchString(song EchoesSong) string {
	// Ignores album for the moment
	searchStrings := []string{
		prependSearchType("track", song.Title),
		prependSearchType("artist", song.Artist)}

	return strings.Join(searchStrings, " ")
}

func prependSearchType(searchType, title string) string {
	return fmt.Sprintf("%v:\"%v\"", searchType, title)
}
