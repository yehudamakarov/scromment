package main

import (
	"flag"
	"log"
	"os"
	"strings"
)

const commentableSplitChar = "|"

func makeOutFile(err error, outLocation string) *os.File {
	out, err := os.Create(outLocation)
	check(err)
	return out
}

func getScriptFile(fileLocation string) (*os.File, error) {
	file, err := os.Open(fileLocation)
	check(err)
	return file, err
}

func closeFile(file *os.File) {
	err := file.Close()
	check(err)
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func parseFlags() (string, string, []string, string) {
	fileLocationPtr := flag.String("file-location", "", "location of the sql file that needs to be handled")
	outLocationPtr := flag.String("out-location", "", "location of the generated sql file (that will be commented appropriately)")
	commentablePtr := flag.String("commentable", "", "text token(s) that must be contained in an object for this tool to comment it out. (1 token - pass a string, 2 or more tokens - pass a pipe (|) delimited string)")
	splitterPtr := flag.String("splitter", "/****** Object:", "text token that dictates when to split objects to be commented out when containing a commentable.")

	flag.Parse()

	fileLocation := *fileLocationPtr
	outLocation := *outLocationPtr
	commentable := *commentablePtr
	splitter := *splitterPtr

	handleCliErrors(fileLocation, outLocation, commentable, splitter)

	commentables := getTokens(commentable)

	return fileLocation, outLocation, commentables, splitter
}

func getTokens(commentable string) []string {
	return strings.Split(commentable, commentableSplitChar)
}

func handleCliErrors(fileLocation string, outLocation string, commentable string, splitter string) {
	// if any arg is missing
	if fileLocation == "" {
		showMissingArgMessage("file-location")
	}
	if outLocation == "" {
		showMissingArgMessage("out-location")
	}
	if commentable == "" {
		showMissingArgMessage("commentable")
	}
	if splitter == "" {
		showMissingArgMessage("splitter")
	}
	// if used a | incorrectly for commentables
	if strings.HasPrefix(commentable, commentableSplitChar) {
		log.Fatal("do not start with a |")
	}
	if strings.HasPrefix(commentable, commentableSplitChar) {
		log.Fatal("do not end with a |")
	}
}

func showMissingArgMessage(missing string) {
	log.Fatalf("don't forget the parameter for: %s!", missing)
}
