package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
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

func playAudio(audioData []byte) {
	// Create a temporary file to store the audio data
	tmpFile, err := os.CreateTemp("", "tts-*.mp3")
	if err != nil {
		catchErr(err)
		return
	}
	defer os.Remove(tmpFile.Name()) // Clean up the temp file when done

	// Write audio data to the temp file
	if _, err := tmpFile.Write(audioData); err != nil {
		tmpFile.Close()
		catchErr(err)
		return
	}
	tmpFile.Close()

	switch runtime.GOOS {
	case "darwin":
		cmd := exec.Command("afplay", tmpFile.Name())
		if err := cmd.Run(); err != nil {
			catchErr(err)
			return
		}
	case "linux":
		cmd := exec.Command("aplay", tmpFile.Name())
		if err := cmd.Run(); err != nil {
			catchErr(err)
			return
		}
	case "windows":
		cmd := exec.Command("start", tmpFile.Name())
		if err := cmd.Run(); err != nil {
			catchErr(err)
			return
		}
	}
}
