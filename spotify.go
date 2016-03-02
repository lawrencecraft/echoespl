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
