package file

import (
	"bufio"
	"bytes"
	"os"
	"strings"
)

func Read(filename string) ([]byte, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func Parse(data []byte) (string, error) {
	buf := bytes.NewBuffer(data)
	fileScanner := bufio.NewScanner(buf)

	var script string
	for fileScanner.Scan() {
		line := fileScanner.Text()

		// Trim comments.
		splits := strings.Split(line, "#")
		if len(splits) < 1 {
			continue
		}
		line = splits[0]

		// Split on whitespace.
		words := strings.Fields(line)
		for i, w := range words {
			script += w
			if i < len(w)-1 {
				script += " "
			}
		}
	}
	return script, nil
}
