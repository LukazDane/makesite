package main

import (
	"container/list"
	"context"
	"crypto/sha1"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2/clientcredentials"
)

// set up structs to hold song/playlist/album info

type Song struct {
	duration int64  //song length in seconds
	title    string //song title
	artist   *Artist
}

type Playlist struct {
	title       string
	description string
	duration    int64
	publishedAt int32
	songs       []*Song
}

type RecentlyPlayed struct {
	capacity int
	size     int
	cache    map[int]*list.Element
	lru      *list.List
}

func NewRecentlyPlayed(capacity int) *RecentlyPlayed {
	rp := RecentlyPlayed{
		capacity: capacity,
		size:     0,
		lru:      list.New(),
		cache:    make(map[string]*list.Element),
	}
	return &rp
}

type Player struct {
	playProgress       int
	RecentlyPlayedList *RecentlyPlayed
}

func NewPlayer() *Player {
	p := Player{
		RecentlyPlayedList: NewRecentlyPlayed(30),
	}
	return &p
}
func (player *Player) Play(playlist *Playlist) {
	if cached, playlist := player.RecentlyPlayedList.Get(playlist.hash()); cached {
		// Play from cache...
	} else {
		// Fetch over the network and start playing...
		player.RecentlyPlayedList.Set(playlist)
	}
}
func (rp *RecentlyPlayed) Get(key string) (*Playlist, bool) {
	if elem, present := rp.cache[key]; present {
		rp.lru.MoveToFront(elem)
		return elem.Value.(*Playlist), true
	} else {
		return nil, false
	}
}
func (playlist *Playlist) hash() string {
	hash := sha1.New()
	s := fmt.Sprintf(
		"%d-%s-%s-%d",
		playlist.duration,
		playlist.title,
		playlist.description,
		playlist.publishedAt,
	)
	hash.Write([]byte(s))
	sum := hash.Sum(nil)
	return fmt.Sprintf("%x", sum)
}
func (rp *RecentlyPlayed) Set(playlist *Playlist) {
	key := playlist.hash()
	if elem, present := rp.cache[key]; present {
		rp.lru.MoveToFront(elem)
	} else {
		elem := rp.lru.PushFront(playlist)
		rp.size++
	}
	rp.cache[key] = elem
}
func (rp *RecentlyPlayed) increment(element *list.Element) {
	rp.lru.MoveToFront(element)
	if rp.size == rp.capacity {
		lruItem := rp.lru.Back()
		rp.lru.Remove(lruItem)
		rp.size--
	}
}
func (rp *RecentlyPlayed) Set(playlist *Playlist) {
	key := playlist.hash()
	if elem, present := rp.cache[key]; present {
		rp.increment(elem) // <- the change
	} else {
		elem := rp.lru.PushFront(playlist)
		rp.size++
	}
	rp.cache[key] = elem
}

func (rp *RecentlyPlayed) Get(key string) (*Playlist, bool) {
	if elem, present := rp.cache[key]; present {
		rp.increment(elem) // <- the change
		return elem.Value.(*Playlist), true
	} else {
		return nil, false
	}
}
func goDotEnvVariable(key string) string {

	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}
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
	results, err := client.Search("Watsky", spotify.SearchTypePlaylist|spotify.SearchTypeAlbum)
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
