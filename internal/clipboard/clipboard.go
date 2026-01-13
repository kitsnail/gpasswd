package clipboard

import (
	"fmt"
	"time"

	"github.com/atotto/clipboard"
)

// Copy copies text to the system clipboard
func Copy(text string) error {
	if err := clipboard.WriteAll(text); err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}
	return nil
}

// Clear clears the clipboard
func Clear() error {
	if err := clipboard.WriteAll(""); err != nil {
		return fmt.Errorf("failed to clear clipboard: %w", err)
	}
	return nil
}

// CopyWithAutoClear copies text to clipboard and clears it after the specified duration
// Returns a channel that will be closed when the clipboard is cleared
func CopyWithAutoClear(text string, duration time.Duration) (<-chan bool, error) {
	if err := Copy(text); err != nil {
		return nil, err
	}

	done := make(chan bool)

	go func() {
		time.Sleep(duration)
		Clear()
		close(done)
	}()

	return done, nil
}

// Get retrieves the current clipboard content
func Get() (string, error) {
	content, err := clipboard.ReadAll()
	if err != nil {
		return "", fmt.Errorf("failed to read clipboard: %w", err)
	}
	return content, nil
}
