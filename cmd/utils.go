package cmd

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/pterm/pterm"
)

var spinner *pterm.SpinnerPrinter
var moonSequence = []string{"🌑 ", "🌒 ", "🌓 ", "🌔 ", "🌕 ", "🌖 ", "🌗 ", "🌘 "}
var ponderSpinner = &pterm.SpinnerPrinter{
	Sequence:            []string{"▀ ", " ▀", " ▄", "▄ "},
	Style:               &pterm.ThemeDefault.SpinnerStyle,
	Delay:               time.Millisecond * 200,
	ShowTimer:           false,
	TimerRoundingFactor: time.Second,
	TimerStyle:          &pterm.ThemeDefault.TimerStyle,
	MessageStyle:        &pterm.ThemeDefault.SpinnerTextStyle,
	InfoPrinter:         &pterm.Info,
	SuccessPrinter:      &pterm.Success,
	FailPrinter:         &pterm.Error,
	WarningPrinter:      &pterm.Warning,
	RemoveWhenDone:      true,
	Text:                "Pondering...",
}

func expanding(emoji string, maxRadius int) []string {
	totalLength := maxRadius*2 - 1          // Total fixed length for each line
	sequence := make([]string, maxRadius*2) // Frames for expanding and contracting

	// Generate expanding sequence
	for i := 0; i < maxRadius; i++ {
		spacesBefore := strings.Repeat(" ", maxRadius-i-1)
		numEmojis := i + 1
		emojis := strings.Repeat(emoji+" ", numEmojis)
		spacesAfterCount := totalLength - len(spacesBefore) - len(emojis)
		if spacesAfterCount < 0 {
			spacesAfterCount = 0 // Ensure spacesAfterCount is never negative
		}
		spacesAfter := strings.Repeat(" ", spacesAfterCount)
		sequence[i] = spacesBefore + emojis + spacesAfter
	}

	// Generate contracting sequence
	for i := 0; i < maxRadius; i++ {
		spacesBefore := strings.Repeat(" ", i+1)
		numEmojis := maxRadius - i
		emojis := strings.Repeat(emoji+" ", numEmojis)
		spacesAfterCount := totalLength - len(spacesBefore) - len(emojis)
		if spacesAfterCount < 0 {
			spacesAfterCount = 0 // Prevent negative space count
		}
		spacesAfter := strings.Repeat(" ", spacesAfterCount)
		sequence[maxRadius+i] = spacesBefore + emojis + spacesAfter
	}

	return sequence
}

func syntaxHighlight(message string) {
	lines := strings.Split(message, "\n")
	var codeBuffer bytes.Buffer
	var inCodeBlock bool
	var currentLexer chroma.Lexer

	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}
	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	// Regex to find inline code and double-quoted text
	backtickRegex := regexp.MustCompile("`([^`]*)`")
	doubleQuoteRegex := regexp.MustCompile(`"([^"]*)"`)
	cyan := "\033[36m"   // Cyan color ANSI escape code
	yellow := "\033[33m" // Yellow color ANSI escape code
	reset := "\033[0m"   // Reset ANSI escape code

	processLine := func(line string) string {
		line = backtickRegex.ReplaceAllStringFunc(line, func(match string) string {
			return cyan + strings.Trim(match, "`") + reset
		})
		line = doubleQuoteRegex.ReplaceAllStringFunc(line, func(match string) string {
			return yellow + match + reset
		})
		return line
	}

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, "```") {
			if inCodeBlock {
				iterator, err := currentLexer.Tokenise(nil, codeBuffer.String())
				if err == nil {
					formatter.Format(os.Stdout, style, iterator)
				}
				fmt.Println()
				codeBuffer.Reset()
				inCodeBlock = false
			} else {
				inCodeBlock = true
				lang := strings.TrimPrefix(trimmedLine, "```")
				currentLexer = lexers.Get(lang)
				if currentLexer == nil {
					currentLexer = lexers.Fallback
				}
			}
		} else if inCodeBlock {
			codeBuffer.WriteString(line + "\n")
		} else {
			fmt.Println("    " + processLine(line))
		}
	}

	if inCodeBlock {
		iterator, err := currentLexer.Tokenise(nil, codeBuffer.String())
		if err == nil {
			formatter.Format(os.Stdout, style, iterator)
		}
		fmt.Println()
	}
}

func fileNameFromURL(urlStr string) string {
	u, err := url.Parse(urlStr)
	catchErr(err)
	// Get the last path component of the URL
	filename := filepath.Base(u.Path)
	// Replace any characters that are not letters, numbers, or underscores with dashes
	filename = regexp.MustCompile(`[^a-zA-Z0-9_]+`).ReplaceAllString(filename, "-")
	// Limit the filename to 255 characters
	if len(filename) >= 255 {
		filename = filename[:255]
	}
	return filename
}

func catchErr(err error, level ...string) {
	if err != nil {
		// Default level is "warn" if none is provided
		lvl := "warn"
		if len(level) > 0 {
			lvl = level[0] // Use the provided level
		}

		fmt.Println("")
		switch lvl {
		case "warn":
			fmt.Println("❗️", err)
		case "fatal":
			fmt.Println("💀", err)
			os.Exit(1)
		}
	}
}

func formatPrompt(prompt string) string {
	// Replace any characters that are not letters, numbers, or underscores with dashes
	return regexp.MustCompile(`[^a-zA-Z0-9_]+`).ReplaceAllString(prompt, "-")
}

func trace() {
	pc := make([]uintptr, 10) // at least 1 entry needed
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	file, line := f.FileLine(pc[0])
	fmt.Printf("%s:%d\n%s\n", file, line, f.Name())
}

// playAudio plays audio from a byte slice.
// func playAudio(audioContent []byte) {
// 	if verbose {
// 		fmt.Println("🔊 Playing audio...")
// 	}

// 	// Create an io.Reader from the byte slice
// 	reader := bytes.NewReader(audioContent)

// 	// Wrap the reader in a NopCloser to make it an io.ReadCloser.
// 	readCloser := io.NopCloser(reader)

// 	// Decode the MP3 stream.
// 	streamer, format, err := mp3.Decode(readCloser)
// 	catchErr(err)
// 	defer streamer.Close()

// 	// Initialize the speaker with the sample rate of the audio and a buffer size.
// 	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
// 	catchErr(err)

// 	// Play the decoded audio.
// 	done := make(chan bool)
// 	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
// 		done <- true
// 	})))

// 	// Wait for the audio to finish playing.
// 	<-done
// }
