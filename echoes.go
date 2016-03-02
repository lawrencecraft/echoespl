package main

import (
	"errors"
	"fmt"
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"io"
	"net/http"
	"strings"
	"time"
)

type EchoesSong struct {
	Title  string
	Album  string
	Artist string
}

func GetShows(uri string) ([]EchoesSong, error) {
	// fetch URL
	client := http.Client{Timeout: 30 * time.Second}
	response, err := client.Get(uri)

	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprint("Error from remote: ", err))
	}

	defer response.Body.Close()
	return getShowsFromStream(response.Body)
}

func getShowsFromStream(stream io.ReadCloser) ([]EchoesSong, error) {
	root, err := html.Parse(stream)

	if err != nil {
		return nil, err
	}

	rows := scrape.FindAll(root, scrape.ByTag(atom.Tr))

	songs := make([]EchoesSong, 0)

	for _, r := range rows {
		song, ok := translateRow(r)
		if ok {
			songs = append(songs, song)
		}
	}

	if len(songs) == 0 {
		return nil, errors.New("Cannot find any songs in document")
	}

	return songs, nil
}

func translateRow(row *html.Node) (EchoesSong, bool) {
	columns := scrape.FindAll(row, scrape.ByTag(atom.Td))
	if len(columns) != 4 {
		return EchoesSong{}, false
	}

	artist := scrape.Text(columns[1])
	if artist == "break" || artist == "Group Name" {
		return EchoesSong{}, false
	}

	s := EchoesSong{}
	s.Artist = scrape.Text(columns[1])
	s.Title = cleanTitle(scrape.Text(columns[2]))
	s.Album = cleanAlbum(scrape.Text(columns[3]))

	return s, true
}

func cleanTitle(title string) string {
	lowercasedCleanedString := strings.Replace(title, "(live)", "", -1)
	cleanedString := strings.Replace(lowercasedCleanedString, "(Live)", "", -1)
	return strings.TrimSpace(cleanedString)
}

func cleanAlbum(album string) string {
	if album == "(unreleased)" {
		return ""
	}

	return album
}
