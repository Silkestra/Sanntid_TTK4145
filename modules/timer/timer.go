package timer

import (
	"fmt"
	"time"
)

type Timer struct {
	duration   time.Duration
	stopChan   chan struct{}
	timeout    chan struct{}
}

// NewTimer initializes and returns a Timer instance.
func NewTimer() *Timer {
	return &Timer{
		stopChan: make(chan struct{}), // Channel to stop the timer
		timeout:  make(chan struct{}), // Channel to signal timeout
	}
}

// Start begins the timer with the specified duration (in seconds).
func (t *Timer) Start(duration float64) {
	t.duration = time.Duration(duration * float64(time.Second))
	go func() {
		select {
		case <-time.After(t.duration):
			close(t.timeout) // Signal that the timer has expired
		case <-t.stopChan:
			return // Timer was stopped before timing out
		}
	}()
}

// Stop deactivates the timer.
func (t *Timer) Stop() {
	close(t.stopChan)
}

// TimedOut checks if the timer has timed out.
func (t *Timer) TimedOut() bool {
	select {
	case <-t.timeout:
		return true
	default:
		return false
	}
}

func main() {
	// Create a new timer instance.
	timer := NewTimer()

	// Start the timer for 5 seconds.
	timer.Start(5.0)

	// Wait until the timer times out.
	for {
		if timer.TimedOut() {
			break
		}
		// You can do something useful here while waiting.
		time.Sleep(100 * time.Millisecond) // Avoid busy-waiting
	}

	// Timer timed out.
	fmt.Println("Timer timed out!")
}
