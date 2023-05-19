package file

import (
	"bufio"
	"bytes"
	"encoding/json"
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

func ParseScript(data []byte) (string, error) {
	buf := bytes.NewBuffer(data)
	fileScanner := bufio.NewScanner(buf)

	var (
		script string
		first  = true
	)
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
		for _, w := range words {
			if !first {
				script += " "
			}

			script += w
			first = false
		}
	}
	return script, nil
}

func ParseTagMap(data []byte) (map[string]string, error) {
	kv := make(map[string]string)
	if err := json.Unmarshal(data, &kv); err != nil {
		return nil, err
	}

	return kv, nil
}
