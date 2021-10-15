package main

import (
	"bufio"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"os"
)

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
