package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("The URL to an Echoes playlist must be provided")
		os.Exit(1)
	}

	url := os.Args[len(os.Args)-1]
	shows, err := GetShows(url)

	if err != nil {
		fmt.Println("There's been a problem: %v", err)
		os.Exit(1)
	}

	for _, show := range shows {
		fmt.Println(show.Title, "|", show.Album, "|", show.Artist)
	}
}
