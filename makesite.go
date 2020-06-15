package main

import (
	"bytes"
	"flag"
	"html/template"
	"io/ioutil"
	"strings"
)

type post struct {
	User    string
	Content string
}

func readFile(fileName string) string {
	fileContents, err := ioutil.ReadFile("first-post.txt")
	if err != nil {
		panic(err)
	}

	return string(fileContents)
}

func renderTemplate(content string) string {
	paths := []string{
		"template.tmpl",
	}

	buff := new(bytes.Buffer)
	t := template.Must(template.New("template.tmpl").ParseFiles(paths...))
	err := t.Execute(buff, post{User: "Luke", Content: content})
	if err != nil {
		panic(err)
	}

	return buff.String()
}

func saveFile(buffer string, fileName string) bool {
	bytesToWrite := []byte(buffer)
	err := ioutil.WriteFile(fileName, bytesToWrite, 0644)

	if err != nil {
		return false
	}

	return true
}

func main() {
	filePtr := flag.String("file", "first-post.txt", "name of file contents to read")
	dirPtr := flag.String("dir", ".", "directory with all files/root")
	flag.Parse()

	content := readFile(*filePtr)
	template := renderTemplate(content)
	fileName := strings.SplitN(*filePtr, ".", 2)[0] + ".html"
	saveFile(template, fileName)
	// fmt.Println(template)
}
