package main

import (
	"os"
	"testing"
)

func TestReaderCanReadSongsFromAnActualDocument(t *testing.T) {
	file, err := os.Open("test_doc.html")

	if err != nil {
		t.Error(err)
	}

	shows, err := getShowsFromStream(file)
	if err != nil {
		t.Error(err)
	}

	found := false
	for _, s := range shows {
		if s.Artist == "Miranda Lee Richards" && s.Title == "It Was Given" && s.Album == "Echoes of the Dreamtime" {
			found = true
		}
	}

	if !found {
		t.Error("Could not find song")
	}
}

func TestReaderCanReadSongsFromAWebsite(t *testing.T) {
	shows, err := GetShows("http://echoes.org/2016/02/23/tuesday-february-23-2016-2/")
	if err != nil {
		t.Error(err)
	}

	found := false
	for _, s := range shows {
		if s.Artist == "John Heart Jackie" && s.Title == "Nevada City" && s.Album == "Episodes" {
			found = true
		}
	}

	if !found {
		t.Error("Could not find song")
	}
}

func TestCleanTitleActuallyRemovesStuffFromTheTitle(t *testing.T) {
	str := cleanTitle("Doobers and Hoosiers (live)")

	if str != "Doobers and Hoosiers" {
		t.Error("Str was", str)
	}
}
