package ui

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/charmbracelet/glamour"
	tsize "github.com/kopoli/go-terminal-size"
)

func ClearScreen() {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	default:
		fmt.Print("\033[H\033[2J")
	}
}

func RenderWithTerminalWidth(raw string) (string, error) {
	s, err := tsize.GetSize()
	width := 0
	if err != nil {
		width = 80
	} else {
		width = s.Width
	}

	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width-5),
	)

	out, err := r.Render(raw)
	if err != nil {
		return "", err
	}
	return out, nil
}
