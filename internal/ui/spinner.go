package ui

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/mattn/go-isatty"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

var Console = &consoleWriter{}

type consoleWriter struct {
	mu    sync.Mutex
	title string // non-empty while a spinner is active
}

func (w *consoleWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.title != "" {
		os.Stderr.WriteString("\r\033[K")
	}
	return os.Stderr.Write(p)
}

func (w *consoleWriter) tick(frame string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.title != "" {
		fmt.Fprintf(os.Stderr, "\r%s %s", frame, w.title)
	}
}

func (w *consoleWriter) setTitle(title string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.title != "" {
		os.Stderr.WriteString("\r\033[K")
	}
	w.title = title
}

func Spin(title string, fn func() error) error {
	if !isatty.IsTerminal(os.Stderr.Fd()) {
		return fn()
	}

	Console.setTitle(title)
	defer Console.setTitle("")

	done := make(chan error, 1)
	go func() { done <- fn() }()

	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()

	for i := 0; ; i++ {
		select {
		case err := <-done:
			return err
		case <-ticker.C:
			Console.tick(Accent.Render(spinnerFrames[i%len(spinnerFrames)]))
		}
	}
}
