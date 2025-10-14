package ui

import (
	"os"
	"strings"
	"unicode/utf8"
)

// Theme defines ANSI escape codes for colors.
type Theme struct {
	Reset  string
	Bold   string
	Gray   string
	Cyan   string
	Blue   string
	Green  string
	Yellow string
	Red    string
	Emoji  bool // toggle emoji display
}

// Predefined themes
var (
	NoColor = Theme{Emoji: true}

	DarkTheme = Theme{
		Reset:  "\033[0m",
		Bold:   "\033[1m",
		Gray:   "\033[90m",
		Cyan:   "\033[36m",
		Blue:   "\033[34m",
		Green:  "\033[32m",
		Yellow: "\033[33m",
		Red:    "\033[31m",
		Emoji:  true,
	}

	LightTheme = Theme{
		Reset:  "\033[0m",
		Bold:   "\033[1m",
		Gray:   "\033[37m",
		Cyan:   "\033[36m",
		Blue:   "\033[34m",
		Green:  "\033[32m",
		Yellow: "\033[33m",
		Red:    "\033[31m",
		Emoji:  true,
	}
)

func SupportsColor() bool {
	term := os.Getenv("TERM")
	if term == "" || strings.Contains(term, "dumb") {
		return false
	}
	return true
}

func SupportsEmoji() bool {
	// simple heuristic: if stdout is UTF-8 capable
	enc := os.Getenv("LANG")
	return utf8.ValidString(enc) && (strings.Contains(strings.ToLower(enc), "utf") || strings.Contains(strings.ToLower(enc), "utf-8"))
}

func GetTheme(colorOpt string, emojiOpt string) Theme {
	var t Theme
	switch strings.ToLower(colorOpt) {
	case "none":
		t = NoColor
	case "light":
		t = LightTheme
	case "dark":
		t = DarkTheme
	case "auto":
		if SupportsColor() {
			t = DarkTheme
		} else {
			t = NoColor
		}
	default:
		t = DarkTheme
	}

	switch strings.ToLower(emojiOpt) {
	case "off":
		t.Emoji = false
	case "on":
		t.Emoji = true
	case "auto":
		t.Emoji = SupportsEmoji()
	default:
		t.Emoji = true
	}

	return t
}
