package routine

import (
	"time"

	"github.com/dohernandez/errors"
)

// GracefulShutdown closes the shutdown channel and waits for the done channels to be closed or exceed the ttl deadline.
func GracefulShutdown(ttl time.Duration, shutdown chan<- struct{}, done ...<-chan struct{}) error {
	close(shutdown)

	return CheckShutdownDeadline(ttl, done...)
}

// CheckShutdownDeadline checks if the shutdown deadline has been exceeded.
func CheckShutdownDeadline(ttl time.Duration, done ...<-chan struct{}) error {
	deadline := time.After(ttl)

	for _, d := range done {
		select {
		case <-d:
			continue
		case <-deadline:
			return errors.New("shutdown deadline exceeded")
		}
	}

	return nil
}
