package single_elevator

import (
	"time"
)

const _pollRate = 20 * time.Millisecond

// Global variables
var (
	timerEndTime          time.Time
	timerActive           bool
	timerEndTimeAvailable time.Time
	timerActiveAvailable  bool
)

// Start the timer with a given duration in seconds
func TimerStart(duration float64, timerType string) {
	switch timerType {
	case "available":
		timerEndTimeAvailable = time.Now().Add(time.Duration(duration) * 3 * time.Second)
		timerActiveAvailable = true
	case "door":
		timerEndTime = time.Now().Add(time.Duration(duration) * time.Second)
		timerActive = true
	}
}

// Stop the timer
func TimerStop(timerType string) {
	switch timerType {
	case "door":
		timerActive = false
	case "available":
		timerActiveAvailable = false
	}
}

// Check if the timer has timed out
func TimerTimedOut() bool {
	//fmt.Println(timerActive, time.Now().After(timerEndTime))
	return timerActive && time.Now().After(timerEndTime) && !ObstructionActive
}
func TimerTimedOutAvailable(elev Elevator) bool {
	//fmt.Println(timerActive, time.Now().After(timerEndTime))
	active_requests := false
	for i := 0; i < 4; i++ {
		for j := 0; j < 3; j++ {
			if elev.Requests[i][j] {
				active_requests = true
				break
			}
		}
	}
	return timerActiveAvailable && time.Now().After(timerEndTimeAvailable) && active_requests
}

func PollAvailableTimeout(receiver chan<- bool, elev *Elevator) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := TimerTimedOutAvailable(*elev)
		if v != prev {
			TimerStop("available")
			receiver <- v
		}
		prev = v
	}
}

func PollTimeout(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := TimerTimedOut()
		if v != prev {
			TimerStop("door")
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
