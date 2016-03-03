package main

import (
	"testing"
)

//func TestPrependSearchTypeSplitsASingleString(t *testing.T) {
//	assert_length(t, prependSearchType("artist", "Chumbawumba"), 1)
//}
//
//func TestPrependSearchTypeSplitsAStringWithSpaces(t *testing.T) {
//	assert_length(t, prependSearchType("artist", "The Artist Formerly Known as Prince"), 6)
//}
//
//func TestPrependSearchTypeReturnsEmptySliceForSpaces(t *testing.T) {
//	assert_length(t, prependSearchType("artist", "       "), 0)
//}
//
//func TestPrependSearchTypePrependsSearchForSingleTerm(t *testing.T) {
//	expected := append(make([]string, 0), "artist:Tubes")
//	actual := prependSearchType("artist", "Tubes")
//
//	assert_equivalent(t, expected, actual)
//}
//
//func TestPrependSearchTypePrependsSearchForMutipleTerm(t *testing.T) {
//	expected := append(make([]string, 0), "artist:Tubes", "artist:Are", "artist:Great")
//	actual := prependSearchType("artist", "Tubes Are Great")
//
//	assert_equivalent(t, expected, actual)
//}
//
//func TestPrependSearchTypeHandlesMultipleSpaces(t *testing.T) {
//	expected := []string{"artist:Tubes", "artist:Are", "artist:Great"}
//	actual := prependSearchType("artist", "Tubes     Are    Great")
//
//	assert_equivalent(t, expected, actual)
//}
//
//func TestEchoesSongToSearchStringProperlyHandlesASong(t *testing.T) {
//	song := EchoesSong{Title: "Tubes Are Great", Artist: "Tubemeister", Album: "Best of Tubes"}
//	expected := "track:Tubes track:Are track:Great artist:Tubemeister album:Best album:of album:Tubes"
//	actual := echoesSongToSearchString(song)
//
//	if actual != expected {
//		t.Errorf("Expected %v but got %v", expected, actual)
//	}
//}
//
//func TestEchoesSongOnlyIncludesFilledOutInput(t *testing.T) {
//	song := EchoesSong{Title: "Tubes Are Great", Artist: "Tubemeister"}
//	expected := "track:Tubes track:Are track:Great artist:Tubemeister"
//	actual := echoesSongToSearchString(song)
//
//	if actual != expected {
//		t.Errorf("Expected %v but got %v", expected, actual)
//	}
//}

func TestEchoesSongsToSearchStringsCreatesAProperLengthArray(t *testing.T) {
	song1 := EchoesSong{Title: "1"}
	song2 := EchoesSong{Title: "22"}
	song3 := EchoesSong{Title: "33"}
	songs := append(make([]EchoesSong, 0), song1, song2, song3)

	terms := echoesSongsToSearchStrings(songs)

	assert_length(t, terms, 3)
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
