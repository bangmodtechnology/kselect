package output

import (
	"fmt"
	"os"
	"sync"
	"time"

	"golang.org/x/term"
)

// Spinner displays an animated loading indicator on stderr.
// It only shows when stderr is a TTY (no output during pipe/redirect).
type Spinner struct {
	message string
	frames  []string
	stop    chan struct{}
	done    sync.WaitGroup
	active  bool
}

// NewSpinner creates a spinner with the given message.
func NewSpinner(message string) *Spinner {
	return &Spinner{
		message: message,
		frames:  []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		stop:    make(chan struct{}),
	}
}

// Start begins the spinner animation in a goroutine.
// Does nothing if stderr is not a TTY.
func (s *Spinner) Start() {
	if !term.IsTerminal(int(os.Stderr.Fd())) {
		return
	}
	s.active = true
	s.done.Add(1)
	go func() {
		defer s.done.Done()
		i := 0
		for {
			select {
			case <-s.stop:
				fmt.Fprintf(os.Stderr, "\r\033[K")
				return
			default:
				fmt.Fprintf(os.Stderr, "\r%s %s", s.frames[i%len(s.frames)], s.message)
				i++
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()
}

// Stop halts the spinner and clears its line.
func (s *Spinner) Stop() {
	if !s.active {
		return
	}
	select {
	case <-s.stop:
	default:
		close(s.stop)
	}
	s.done.Wait()
}
