package main

import (
	"bufio"
	"os"
	"strings"
)

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
