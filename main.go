package main

import (
	"bufio"
	"flag"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"log"
	"os"
	"strings"
)

func main() {
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

	file, err := os.Open(fileLocation)
	check(err)
	defer func(file *os.File) {
		err := file.Close()
		check(err)
	}(file)

	out, err := os.Create(outLocation)
	check(err)
	defer func(out *os.File) {
		err := out.Close()
		check(err)
	}(out)

	rewriteFile(file, splitter, commentable, out)

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
	log.Fatalf("don't forget %s!", missing)
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

func rewriteFile(file *os.File, splitter string, commentable string, out *os.File) {
	decoder := getDecoder(file)
	scanner := bufio.NewScanner(transform.NewReader(file, decoder))
	iterate(scanner, splitter, commentable, out)
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

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
