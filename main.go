package main

import (
	"bufio"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"os"
	"strings"
)

func main() {
	fileLocation, outLocation, commentable, splitter := parseFLags()

	file, err := getScriptFile(fileLocation)
	defer closeFile(file)

	out := makeOutFile(err, outLocation)
	defer closeFile(out)

	editFile(file, splitter, commentable, out)
}

type chunk struct {
	shouldComment bool
	header        string
	sb            *strings.Builder
}

func (c *chunk) addCurrentLineToChunk(s string) {
	c.sb.WriteString(s)
	c.sb.WriteString("\n")
}

func (c *chunk) weShouldCommentThisChunk() {
	c.shouldComment = true
}

func weShouldCommentThisLine(s string, commentable string) bool {
	containsCommentable := strings.Contains(strings.ToLower(s), strings.ToLower(commentable))
	notAlreadyCommentedOut := !strings.HasPrefix(strings.TrimSpace(s), "--")
	return containsCommentable && notAlreadyCommentedOut
}

func editFile(file *os.File, splitter string, commentable string, out *os.File) {
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

func getFileContent(file *os.File) *bufio.Scanner {
	decoder := getDecoder(file)
	scanner := bufio.NewScanner(transform.NewReader(file, decoder))
	return scanner
}

func getDecoder(file *os.File) transform.Transformer {
	b := make([]byte, 2)
	_, err := file.Read(b)
	check(err)
	var decoder transform.Transformer
	if (b[0] == 255 && b[1] == 254) || (b[0] == 254 && b[1] == 255) {
		decoder = unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder()
	} else {
		decoder = unicode.UTF8.NewDecoder()
	}
	_, err = file.Seek(0, 0)
	check(err)
	return decoder
}

func (c *chunk) writeChunk(f *os.File) {
	if c.sb.Len() == 0 {
		return
	}
	var toWrite strings.Builder

	// prep header
	if c.shouldComment {
		toWrite.WriteString("-- Commented by scromment ----- \n")
	}
	toWrite.WriteString(c.header + "\n")

	// prep body
	content := c.sb.String()
	rescan := bufio.NewScanner(strings.NewReader(content))
	for rescan.Scan() {
		if c.shouldComment {
			toWrite.WriteString("--")
		}
		toWrite.WriteString(rescan.Text())
		toWrite.WriteString("\n")
	}

	// write
	_, err := f.WriteString(toWrite.String())
	check(err)
}
