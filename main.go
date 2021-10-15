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

func (c *chunk) addContent(s string, commentable string) {
	if strings.Contains(strings.ToLower(s), strings.ToLower(commentable)) && !strings.HasPrefix(strings.TrimSpace(s), "--") {
		c.shouldComment = true
	}
	c.sb.WriteString(s)
	c.sb.WriteString("\n")
}

func editFile(file *os.File, splitter string, commentable string, out *os.File) {
	scanner := getFileContent(file)
	iterate(scanner, splitter, commentable, out)
}

func getFileContent(file *os.File) *bufio.Scanner {
	decoder := getDecoder(file)
	scanner := bufio.NewScanner(transform.NewReader(file, decoder))
	return scanner
}

func iterate(scanner *bufio.Scanner, splitter string, commentable string, out *os.File) {
	c := chunk{sb: &strings.Builder{}}
	scanner.Scan()
	firstLine := scanner.Text()
	if strings.HasPrefix(firstLine, splitter) {
		c.header = firstLine
	} else {
		c.addContent(firstLine, commentable)
	}

	for scanner.Scan() {
		line := scanner.Text()
		if encounteredHeader := strings.HasPrefix(line, splitter); encounteredHeader {
			writeChunk(c, out)
			c = chunk{sb: &strings.Builder{}, header: line}
		} else {
			c.addContent(line, commentable)
		}
	}

	// write contents after last header
	if c.sb.Len() > 0 {
		writeChunk(c, out)
	}
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

func writeChunk(c chunk, f *os.File) {
	var toWrite strings.Builder
	if c.shouldComment {
		toWrite.WriteString("-- Commented by scromment ----- ")
	}
	toWrite.WriteString(c.header + "\n")
	content := c.sb.String()
	rescan := bufio.NewScanner(strings.NewReader(content))
	for rescan.Scan() {
		if c.shouldComment {
			toWrite.WriteString("--")
		}
		toWrite.WriteString(rescan.Text())
		toWrite.WriteString("\n")
	}
	_, err := f.WriteString(toWrite.String())
	check(err)
	c.sb.Reset()
}
