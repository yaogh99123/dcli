package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	// "github.com/jesseduffield/yaml"

	"github.com/goccy/go-yaml"
)

// SplitLines takes a multiline string and splits it on newlines
// currently we are also stripping \r's which may have adverse effects for
// windows users (but no issues have been raised yet)
func SplitLines(multilineString string) []string {
	multilineString = strings.Replace(multilineString, "\r", "", -1)
	if multilineString == "" || multilineString == "\n" {
		return make([]string, 0)
	}
	lines := strings.Split(multilineString, "\n")
	if lines[len(lines)-1] == "" {
		return lines[:len(lines)-1]
	}
	return lines
}

// NormalizeLinefeeds - Removes all Windows and Mac style line feeds
func NormalizeLinefeeds(str string) string {
	str = strings.Replace(str, "\r\n", "\n", -1)
	str = strings.Replace(str, "\r", "", -1)
	return str
}

// Loader dumps a string to be displayed as a loader
func Loader() string {
	characters := "|/-\\"
	now := time.Now()
	nanos := now.UnixNano()
	index := nanos / 50000000 % int64(len(characters))
	return characters[index : index+1]
}

// ResolvePlaceholderString populates a template with values
func ResolvePlaceholderString(str string, arguments map[string]string) string {
	for key, value := range arguments {
		str = strings.Replace(str, "{{"+key+"}}", value, -1)
	}
	return str
}

// Max returns the maximum of two integers
func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// WithShortSha returns a command but with a shorter SHA. in the terminal we're all used to 10 character SHAs but under the hood they're actually 64 characters long. No need including all the characters when we're just displaying a command
func WithShortSha(str string) string {
	split := strings.Split(str, " ")
	for i, word := range split {
		// good enough proxy for now
		if len(word) == 64 {
			split[i] = word[0:10]
		}
	}
	return strings.Join(split, " ")
}

type multiErr []error

func (m multiErr) Error() string {
	var b bytes.Buffer
	b.WriteString("encountered multiple errors:")
	for _, err := range m {
		b.WriteString("\n\t... " + err.Error())
	}
	return b.String()
}

func CloseMany(closers []io.Closer) error {
	errs := make([]error, 0, len(closers))
	for _, c := range closers {
		err := c.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return multiErr(errs)
	}
	return nil
}

func SafeTruncate(str string, limit int) string {
	if len(str) > limit {
		return str[0:limit]
	} else {
		return str
	}
}

func IsValidHexValue(v string) bool {
	if len(v) != 4 && len(v) != 7 {
		return false
	}

	if v[0] != '#' {
		return false
	}

	for _, char := range v[1:] {
		switch char {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f', 'A', 'B', 'C', 'D', 'E', 'F':
			continue
		default:
			return false
		}
	}

	return true
}

// MarshalIntoYaml gets any json-tagged data and marshal it into yaml saving original json structure.
// Useful for structs from 3rd-party libs without yaml tags.
func MarshalIntoYaml(data interface{}) ([]byte, error) {
	return marshalIntoFormat(data, "yaml")
}

func marshalIntoFormat(data interface{}, format string) ([]byte, error) {
	// First marshal struct->json to get the resulting structure declared by json tags
	dataJSON, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, err
	}
	switch format {
	case "json":
		return dataJSON, err
	case "yaml":
		// Use Unmarshal->Marshal hack to convert json into yaml with the original structure preserved
		var dataMirror yaml.MapSlice
		if err := yaml.Unmarshal(dataJSON, &dataMirror); err != nil {
			return nil, err
		}
		return yaml.Marshal(dataMirror)
	default:
		return nil, fmt.Errorf("Unsupported detailization format: %s", format)
	}
}
