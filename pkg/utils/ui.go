package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"math"
	"regexp"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/goccy/go-yaml/lexer"
	"github.com/goccy/go-yaml/printer"
	"github.com/go-errors/errors"
	"github.com/jesseduffield/gocui"
	"github.com/mattn/go-runewidth"
)

// 颜色定义 (标准 ANSI 颜色)
const (
	ColorRed     = "\033[0;31m"
	ColorGreen   = "\033[0;32m"
	ColorYellow  = "\033[1;33m"
	ColorBlue    = "\033[0;34m"
	ColorMagenta = "\033[0;35m"
	ColorCyan    = "\033[0;36m"
	ColorWhite   = "\033[0;37m"
	ColorNC      = "\033[0m" // No Color
)

// ColoredString takes a string and a colour attribute and returns a colored
// string with that attribute
func ColoredString(str string, colorAttribute color.Attribute) string {
	if colorAttribute == color.FgWhite {
		return str
	}
	colour := color.New(colorAttribute)
	return ColoredStringDirect(str, colour)
}

// ColoredStringDirect used for aggregating a few color attributes rather than
// just sending a single one
func ColoredStringDirect(str string, colour *color.Color) string {
	return colour.SprintFunc()(fmt.Sprint(str))
}

// MultiColoredString takes a string and an array of colour attributes and returns a colored
// string with those attributes
func MultiColoredString(str string, colorAttribute ...color.Attribute) string {
	colour := color.New(colorAttribute...)
	return ColoredStringDirect(str, colour)
}

// Decolorise strips a string of color
func Decolorise(str string) string {
	re := regexp.MustCompile(`\x1B\[([0-9]{1,2}(;[0-9]{1,2})?)?[mK]`)
	return re.ReplaceAllString(str, "")
}

// WithPadding pads a string as much as you want
func WithPadding(str string, padding int) string {
	uncoloredStr := Decolorise(str)
	if padding < runewidth.StringWidth(uncoloredStr) {
		return str
	}
	return str + strings.Repeat(" ", padding-runewidth.StringWidth(uncoloredStr))
}

// RenderTable takes an array of string arrays and returns a table containing the values
func RenderTable(rows [][]string) (string, error) {
	if len(rows) == 0 {
		return "", nil
	}
	if !displayArraysAligned(rows) {
		return "", errors.New("Each item must return the same number of strings to display")
	}

	columnPadWidths := getPadWidths(rows)
	paddedDisplayRows := getPaddedDisplayStrings(rows, columnPadWidths)

	return strings.Join(paddedDisplayRows, "\n"), nil
}

func getPadWidths(rows [][]string) []int {
	if len(rows[0]) <= 1 {
		return []int{}
	}
	columnPadWidths := make([]int, len(rows[0])-1)
	for i := range columnPadWidths {
		for _, cells := range rows {
			uncoloredCell := Decolorise(cells[i])

			if runewidth.StringWidth(uncoloredCell) > columnPadWidths[i] {
				columnPadWidths[i] = runewidth.StringWidth(uncoloredCell)
			}
		}
	}
	return columnPadWidths
}

func getPaddedDisplayStrings(rows [][]string, columnPadWidths []int) []string {
	paddedDisplayRows := make([]string, len(rows))
	for i, cells := range rows {
		for j, columnPadWidth := range columnPadWidths {
			paddedDisplayRows[i] += WithPadding(cells[j], columnPadWidth) + " "
		}
		paddedDisplayRows[i] += cells[len(columnPadWidths)]
	}
	return paddedDisplayRows
}

func displayArraysAligned(stringArrays [][]string) bool {
	for _, strings := range stringArrays {
		if len(strings) != len(stringArrays[0]) {
			return false
		}
	}
	return true
}

// ColoredYamlString takes an YAML formatted string and returns a colored string
func ColoredYamlString(str string) string {
	format := func(attr color.Attribute) string {
		return fmt.Sprintf("%s[%dm", "\x1b", attr)
	}
	tokens := lexer.Tokenize(str)
	var p printer.Printer
	p.Bool = func() *printer.Property {
		return &printer.Property{
			Prefix: format(color.FgMagenta),
			Suffix: format(color.Reset),
		}
	}
	p.Number = func() *printer.Property {
		return &printer.Property{
			Prefix: format(color.FgYellow),
			Suffix: format(color.Reset),
		}
	}
	p.MapKey = func() *printer.Property {
		return &printer.Property{
			Prefix: format(color.FgCyan),
			Suffix: format(color.Reset),
		}
	}
	p.String = func() *printer.Property {
		return &printer.Property{
			Prefix: format(color.FgGreen),
			Suffix: format(color.Reset),
		}
	}
	return p.PrintTokens(tokens)
}

// FormatBinaryBytes formats bytes to binary units (kiB, MiB, etc.)
func FormatBinaryBytes(b int) string {
	n := float64(b)
	units := []string{"B", "kiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB"}
	for _, unit := range units {
		if n > math.Pow(2, 10) {
			n /= math.Pow(2, 10)
		} else {
			val := fmt.Sprintf("%.2f%s", n, unit)
			if val == "0.00B" {
				return "0B"
			}
			return val
		}
	}
	return "a lot"
}

// FormatDecimalBytes formats bytes to decimal units (kB, MB, etc.)
func FormatDecimalBytes(b int) string {
	n := float64(b)
	units := []string{"B", "kB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
	for _, unit := range units {
		if n > math.Pow(10, 3) {
			n /= math.Pow(10, 3)
		} else {
			val := fmt.Sprintf("%.2f%s", n, unit)
			if val == "0.00B" {
				return "0B"
			}
			return val
		}
	}
	return "a lot"
}

// ApplyTemplate populates a template with an object
func ApplyTemplate(str string, object interface{}) string {
	var buf bytes.Buffer
	_ = template.Must(template.New("").Parse(str)).Execute(&buf, object)
	return buf.String()
}

// GetGocuiAttribute gets the gocui color attribute from the string
func GetGocuiAttribute(key string) gocui.Attribute {
	colorMap := map[string]gocui.Attribute{
		"default":   gocui.ColorDefault,
		"black":     gocui.ColorBlack,
		"red":       gocui.ColorRed,
		"green":     gocui.ColorGreen,
		"yellow":    gocui.ColorYellow,
		"blue":      gocui.ColorBlue,
		"magenta":   gocui.ColorMagenta,
		"cyan":      gocui.ColorCyan,
		"white":     gocui.ColorWhite,
		"bold":      gocui.AttrBold,
		"reverse":   gocui.AttrReverse,
		"underline": gocui.AttrUnderline,
	}
	value, present := colorMap[key]
	if present {
		return value
	}
	return gocui.ColorDefault
}

// GetColorAttribute gets the color attribute from the string
func GetColorAttribute(key string) color.Attribute {
	colorMap := map[string]color.Attribute{
		"default":   color.FgWhite,
		"black":     color.FgBlack,
		"red":       color.FgRed,
		"green":     color.FgGreen,
		"yellow":    color.FgYellow,
		"blue":      color.FgBlue,
		"magenta":   color.FgMagenta,
		"cyan":      color.FgCyan,
		"white":     color.FgWhite,
		"bold":      color.Bold,
		"underline": color.Underline,
	}
	value, present := colorMap[key]
	if present {
		return value
	}
	return color.FgWhite
}

// FormatMapItem is for displaying items in a map
func FormatMapItem(padding int, k string, v interface{}) string {
	return fmt.Sprintf("%s%s %v\n", strings.Repeat(" ", padding), ColoredString(k+":", color.FgYellow), fmt.Sprintf("%v", v))
}

// FormatMap is for displaying a map
func FormatMap(padding int, m map[string]string) string {
	if len(m) == 0 {
		return "none\n"
	}

	output := "\n"

	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		output += FormatMapItem(padding, key, m[key])
	}

	return output
}

// OpensMenuStyle style used on menu items that open another menu
func OpensMenuStyle(str string) string {
	return ColoredString(fmt.Sprintf("%s...", str), color.FgMagenta)
}
