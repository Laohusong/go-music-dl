package core

import (
	"regexp"
	"strings"
)

var (
	lineLyricTimestamp = regexp.MustCompile(`^\[\d+:\d{2}(?:\.\d{1,3})?\]`)
	wordLyricTimestamp = regexp.MustCompile(`\[\d+:\d{2}(?:\.\d{1,3})?\]|<\d+:\d{2}(?:\.\d{1,3})?>`)
)

// LineSyncedLyrics removes inline word cues while preserving line start times,
// metadata, translations, repeated-line timestamps and unsynchronised text.
func LineSyncedLyrics(raw string) string {
	lines := strings.Split(strings.TrimPrefix(raw, "\ufeff"), "\n")
	for i, rawLine := range lines {
		line := strings.TrimSuffix(rawLine, "\r")
		remaining := strings.TrimLeft(line, " \t")
		prefix := ""
		for {
			stamp := lineLyricTimestamp.FindString(remaining)
			if stamp == "" {
				break
			}
			prefix += stamp
			remaining = remaining[len(stamp):]
		}
		if prefix != "" {
			lines[i] = prefix + wordLyricTimestamp.ReplaceAllString(remaining, "")
		} else {
			lines[i] = line
		}
	}
	return strings.Join(lines, "\n")
}
