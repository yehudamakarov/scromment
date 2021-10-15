package main

import (
	"bufio"
	"os"
	"strings"
)

func main() {
	fileLocation, outLocation, commentable, splitter := parseFLags()

	file, err := getScriptFile(fileLocation)
	defer closeFile(file)

	out := makeOutFile(err, outLocation)
	defer closeFile(out)

	convertFile(file, splitter, commentable, out)
}

func weShouldCommentThisLine(s string, commentable string) bool {
	containsCommentable := strings.Contains(strings.ToLower(s), strings.ToLower(commentable))
	notAlreadyCommentedOut := !strings.HasPrefix(strings.TrimSpace(s), "--")
	return containsCommentable && notAlreadyCommentedOut
}

func convertFile(file *os.File, splitter string, commentable string, out *os.File) {
	scanner := getFileContent(file)
	iterateAndRewrite(scanner, splitter, commentable, out)
}

func iterateAndRewrite(scanner *bufio.Scanner, splitter string, commentable string, out *os.File) {
	c := chunk{sb: &strings.Builder{}}
	for scanner.Scan() {
		line := scanner.Text()
		if encounteredHeader := strings.HasPrefix(line, splitter); encounteredHeader {
			c.writeChunk(out)
			c = chunk{sb: &strings.Builder{}, header: line}
		} else {
			if weShouldCommentThisLine(line, commentable) {
				c.weShouldCommentThisChunk()
			}
			c.addCurrentLineToChunk(line)
		}
	}
	// writes contents after the last header
	c.writeChunk(out)
}
