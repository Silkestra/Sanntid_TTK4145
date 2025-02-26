package single_elevator

import (
	"time"
)

const _pollRate = 20 * time.Millisecond

// Global variables
var (
	timerEndTime time.Time
	timerActive  bool
)

// Start the timer with a given duration in seconds
func TimerStart(duration float64) {
	timerEndTime = time.Now().Add(time.Duration(duration) * time.Second)
	timerActive = true
}

// Stop the timer
func TimerStop() {
	timerActive = false
}

// Check if the timer has timed out
func TimerTimedOut() bool {
	//fmt.Println(timerActive, time.Now().After(timerEndTime))
	return timerActive && time.Now().After(timerEndTime) && !ObstructionActive
}

func PollTimeout(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := TimerTimedOut()
		if v != prev {
			TimerStop()
			receiver <- v
		}
		prev = v
	}
}

// package timer

// import (
// 	"time"
// )

// type Timer struct {
// 	duration   time.Duration
// 	stopChan   chan struct{}
// 	timeout    chan struct{}
// }

// // NewTimer initializes and returns a Timer instance.
// func NewTimer() *Timer {
// 	return &Timer{
// 		stopChan: make(chan struct{}), // Channel to stop the timer
// 		timeout:  make(chan struct{}), // Channel to signal timeout
// 	}
// }

// // Start begins the timer with the specified duration (in seconds).
// func (t *Timer) Start(duration float64) {
// 	t.duration = time.Duration(duration * float64(time.Second))
// 	go func() {
// 		select {
// 		case <-time.After(t.duration):
// 			close(t.timeout) // Signal that the timer has expired
// 		case <-t.stopChan:
// 			return // Timer was stopped before timing out
// 		}
// 	}()
// }

// // Stop deactivates the timer.
// func (t *Timer) Stop() {
// 	close(t.stopChan)
// }

// // TimedOut checks if the timer has timed out.
// func (t *Timer) TimedOut() bool {
// 	select {
// 	case <-t.timeout:
// 		return true
// 	default:
// 		return false
// 	}
// }

// /* func main() {
// 	// Create a new timer instance.
// 	timer := NewTimer()

// 	// Start the timer for 5 seconds.
// 	timer.Start(5.0)

// 	// Wait until the timer times out.
// 	for {
// 		if timer.TimedOut() {
// 			break
// 		}
// 		// You can do something useful here while waiting.
// 		time.Sleep(100 * time.Millisecond) // Avoid busy-waiting
// 	}

// 	// Timer timed out.
// 	fmt.Println("Timer timed out!")
// }
//  */
