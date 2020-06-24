package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"golang.org/x/oauth2/clientcredentials"

	"github.com/zmb3/spotify"
)

// type post struct {
// 	User    string
// 	Content string
// }

func main() {
	filePtr := flag.String("file", "", "name of file contents to read")
	dirPtr := flag.String("dir", ".", "directory with all files/root")
	flag.Parse()

	if *filePtr != "" {
		makePost(*filePtr)
	} else {
		parseDir(*dirPtr)
	}
	// content := readFile(*filePtr)
	// template := renderTemplate(content)
	// fileName := strings.SplitN(*filePtr, ".", 2)[0] + ".html"
	// saveFile(template, fileName)
	// fmt.Println(template)

	config := &clientcredentials.Config{
		ClientID:     os.Getenv("SPOTIFY_ID"),
		ClientSecret: os.Getenv("SPOTIFY_SECRET"),
		TokenURL:     spotify.TokenURL,
	}
	token, err := config.Token(context.Background())
	if err != nil {
		log.Fatalf("couldn't get token: %v", err)
	}

	client := spotify.Authenticator{}.NewClient(token)
	// search for playlists and albums containing "holiday"
	results, err := client.Search("holiday", spotify.SearchTypePlaylist|spotify.SearchTypeAlbum)
	if err != nil {
		log.Fatal(err)
	}

	// handle album results
	if results.Albums != nil {
		fmt.Println("Albums:")
		for _, item := range results.Albums.Albums {
			fmt.Println("   ", item.Name)
		}
	}
	// handle playlist results
	if results.Playlists != nil {
		fmt.Println("Playlists:")
		for _, item := range results.Playlists.Playlists {
			fmt.Println("   ", item.Name)
		}
	}
}

func parseDir(dir string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Print("Error reading files: ")
		fmt.Println(err)
	} else {
		for _, f := range files {
			if f.IsDir() {
				parseDir(fmt.Sprintf("%s/%s", dir, f.Name()))
			} else if strings.HasSuffix(f.Name(), ".txt") {
				fmt.Println(f.Name())
				fmt.Println(dir + "/" + f.Name())
				makePost(dir + "/" + f.Name())
			}
		}
	}
}

func makePost(name string) {
	content := readFile(name)
	newName := strings.Split(name, ".txt")[0] + ".html"
	renderTemplate(newName, content)
}

func readFile(fileName string) string {
	fileContents, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	return string(fileContents)
}

func renderTemplate(fileName string, text string) {
	paths := []string{
		"template.tmpl",
	}

	t := template.Must(template.New("template.tmpl").ParseFiles(paths...))
	file, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}

	err = t.Execute(file, text)
	if err != nil {
		panic(err)
	}
}

func saveFile(buffer string, fileName string) bool {
	bytesToWrite := []byte(buffer)
	err := ioutil.WriteFile(fileName, bytesToWrite, 0644)

	if err != nil {
		return false
	}

	return true
}

// func checkFlags(name string) bool {
// 	dirFlag := false
// 	flag.Visit(func(f *flag.Flag) {
// 		if f.Name == name {
// 			dirFlag = true
// 		}
// 	})

// 	return dirFlag
// }
