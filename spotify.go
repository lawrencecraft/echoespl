package main

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"github.com/zmb3/spotify"
	"math/rand"
	"net/http"
	"os"
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

// StartAuthentication begins an authentication request to Spotify. It returns
// an authentication URL and a channel. The channel will publish a the token when
// the http server receives the callback.
func StartAuthenticationFlow(clientID string, secretKey string) (string, chan string, spotify.Authenticator) {
	auth := GetDefaultAuthenticator(clientID, secretKey)
	state := generateState()
	authUrl := auth.AuthURL(state)

	replyChannel := make(chan string, 1)

	http.HandleFunc(redirect_url_part, func(w http.ResponseWriter, r *http.Request) {
		transferredState, stateOk := r.URL.Query()["state"]
		code, codeOk := r.URL.Query()["code"]

		if !stateOk || !codeOk || transferredState[0] != state {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Not quite what we're looking for"))
		}

		replyChannel <- code[0]

		w.Write([]byte("All done. Close your browser now."))
	})

	go http.ListenAndServe(redirect_port, nil)

	return authUrl, replyChannel, auth
}

func generateState() string {
	now := time.Now().UTC().UnixNano()
	rand.Seed(now + int64(os.Getpid()))
	compiledString := []byte(fmt.Sprintf(string(now), os.Getpid(), rand.Int()))
	hash := md5.New().Sum(compiledString)
	return base64.StdEncoding.EncodeToString(hash)
}
