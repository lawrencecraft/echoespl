package main

import (
	"testing"
	"time"
)

func TestEchoesSongToSearchStringDoesNotIncludeAlbum(t *testing.T) {
	song := EchoesSong{Title: "Tubes Are Great", Artist: "Tubemeister", Album: "Best of Tubes"}
	expected := "track:\"Tubes Are Great\" artist:\"Tubemeister\""
	actual := echoesSongToSearchString(song)

	if actual != expected {
		t.Errorf("Expected %v but got %v", expected, actual)
	}
}

func TestEchoesSongToSearchStringProperlyHandlesASong(t *testing.T) {
	song := EchoesSong{Title: "Tubes Are Great", Artist: "Tubemeister"}
	expected := "track:\"Tubes Are Great\" artist:\"Tubemeister\""
	actual := echoesSongToSearchString(song)

	if actual != expected {
		t.Errorf("Expected %v but got %v", expected, actual)
	}
}

func TestEchoesSongsToSearchStringsCreatesAProperLengthArray(t *testing.T) {
	song1 := EchoesSong{Title: "11"}
	song2 := EchoesSong{Title: "22"}
	song3 := EchoesSong{Title: "33"}
	songs := append(make([]EchoesSong, 0), song1, song2, song3)

	terms := echoesSongsToSearchStrings(songs)

	assert_length(t, terms, 3)
}

func TestGeneratePlaylistNamePreservesUnixTime(t *testing.T) {
	unixTime := time.Unix(1457039693, 0)
	playlistName := generatePlaylistName(unixTime)

	if playlistName != "Echoespl Playlist 1457039693" {
		t.Error("Expected Echoespl Playlist 1457039693 but got", playlistName)
	}
}

func assert_length(t *testing.T, array []string, expected_length int) {
	if len(array) != expected_length {
		t.Error("Expected length", expected_length, "but instead got", len(array))
	}
}

func assert_equivalent(t *testing.T, expected []string, actual []string) {
	if len(expected) != len(actual) {
		t.Error("Expected length", len(expected), "but actual length", len(actual))
		return
	}

	for i := range actual {
		expectedValue := expected[i]
		actualValue := actual[i]

		if expectedValue != actualValue {
			t.Errorf("At index %v: expected was %v and actual was %v", i, expectedValue, actualValue)
		}
	}
}
