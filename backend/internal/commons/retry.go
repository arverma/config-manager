package commons

import (
	"log"
	"time"
)

// RetryWithBackoff runs fn up to maxAttempts times, with exponential backoff between failures.
// It returns the last error if all attempts fail.
func RetryWithBackoff(maxAttempts int, initialBackoff time.Duration, fn func() error) error {
	var lastErr error
	backoff := initialBackoff
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		lastErr = fn()
		if lastErr == nil {
			return nil
		}
		if attempt == maxAttempts {
			return lastErr
		}
		log.Printf("Connection attempt %d/%d failed, retrying in %v: %v", attempt, maxAttempts, backoff, lastErr)
		time.Sleep(backoff)
		backoff *= 2
	}
	return lastErr
}
