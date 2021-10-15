package main

import (
	"flag"
	"log"
	"os"
)

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

func parseFLags() (string, string, string, string) {
	fileLocationPtr := flag.String("file-location", "", "location of the sql file that needs to be handled")
	outLocationPtr := flag.String("out-location", "", "location of the generated sql file (that will be commented appropriately)")
	commentablePtr := flag.String("commentable", "", "text token that must be contained in an object for this tool to comment it out")
	splitterPtr := flag.String("splitter", "/****** Object:", "text token that dictates when to split objects to be commented out when containing a commentable.")

	flag.Parse()

	fileLocation := *fileLocationPtr
	outLocation := *outLocationPtr
	commentable := *commentablePtr
	splitter := *splitterPtr

	handleCliErrors(fileLocation, outLocation, commentable, splitter)
	return fileLocation, outLocation, commentable, splitter
}

func handleCliErrors(fileLocation string, outLocation string, commentable string, splitter string) {
	if fileLocation == "" {
		showErrorMessage("file-location")
	}
	if outLocation == "" {
		showErrorMessage("out-location")
	}
	if commentable == "" {
		showErrorMessage("commentable")
	}
	if splitter == "" {
		showErrorMessage("splitter")
	}
}

func showErrorMessage(missing string) {
	log.Fatalf("don't forget the parameter for: %s!", missing)
}
