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
	splitterPtr := flag.String("splitter", "/****** Object:", "text token that dictates when to split objects to be commented out when containing a commentable. \n defaults to: /****** Object: for sql server's exported scripts")

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
	msg := "please specify the | "
	shouldThrow := false
	if fileLocation == "" {
		msg = msg + "file location "
		shouldThrow = true
	}
	if outLocation == "" {
		msg = msg + "| out file location "
		shouldThrow = true
	}
	if commentable == "" {
		msg = msg + "| commentable "
		shouldThrow = true
	}
	if splitter == "" {
		msg = msg + "| splitter "
		shouldThrow = true
	}
	msg = msg + "argument(s)"
	if shouldThrow {
		log.Fatal(msg)
	}
}

func rewriteFile(file *os.File, splitter string, commentable string, out *os.File) {

	b := make([]byte, 2)
	_, err := file.Read(b)
	check(err)
	var decoder transform.Transformer
	if (b[0] == 255 && b[1] == 254) || (b[0] == 254 && b[1] == 255) {
		decoder = unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder()
	} else {
		decoder = unicode.UTF8.NewDecoder()
	}

	scanner := bufio.NewScanner(transform.NewReader(file, decoder))
	check(err)
	sb := strings.Builder{}
	var currentHeader string
	for scanner.Scan() {
		line := scanner.Text()

		if encounteredHeader := strings.HasPrefix(line, splitter); encounteredHeader {
			if sb.Len() > 0 {
				handleChunkedObject(sb, commentable, currentHeader, out)
				sb.Reset()
			}
			currentHeader = line + "\n"
		} else {
			sb.WriteString(line + "\n")
		}
	}
	if sb.Len() > 0 {
		handleChunkedObject(sb, commentable, currentHeader, out)
	}
}

func handleChunkedObject(sb strings.Builder, commentable string, currentHeader string, out *os.File) {
	content := sb.String()
	needsCommenting := strings.Contains(strings.ToLower(content), strings.ToLower(commentable))

	if needsCommenting {
		currentHeader = "--Needed Commenting ----- " + currentHeader
	}
	_, err := out.WriteString(currentHeader)
	check(err)

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		text := scanner.Text() + "\n"
		if needsCommenting {
			text = "--" + text
		}
		_, err := out.WriteString(text)
		check(err)
	}
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
