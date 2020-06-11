package alert

import (
	"fmt"
	"strings"
	"time"

	"github.com/gookit/color"
)

type State uint32

// Initialize states in order to simulate a Finite State Machine
// Av ->(<80%/Down) -> Un
// Un ->(>80%/Up) -> Av
const (
	// Normal State -> Available > 80% Availability
	Available State = iota

	// Unavailable -> Over the last 2 minutes, availability is below 80%
	Unavailable
)

type Alert struct {

	// The state of the alert (according to the FSM logic)
	AlertState State

	// The threshold of availability
	Threshold float64

	// Stores the availability of the website
	Availability float64

	// Slice of time.Time that the website went from Down to Up
	// It is initialized with the time the system started (went up for the first time)
	LastTimeAvailable []time.Time

	// Slice of time.Time that the website went from Up to Down
	LastTimeUnavailable []time.Time
}

func NewAlert(t float64) *Alert {
	al := &Alert{
		AlertState:          Available,
		Threshold:           t,
		LastTimeAvailable:   make([]time.Time, 0),
		LastTimeUnavailable: make([]time.Time, 0),
	}
	al.LastTimeAvailable = append(al.LastTimeAvailable, time.Now())
	return al
}

// Function in order to test the functionality of the alert
func (a *Alert) PrintTest() string {
	red := color.FgRed.Render
	green := color.FgGreen.Render
	var res strings.Builder
	if len(a.LastTimeAvailable) == len(a.LastTimeUnavailable) {
		res.WriteString(fmt.Sprintf(red("STATUS: DOWN, Availability: %v, Since: %v, Duration: %v\n"), a.Availability, a.LastTimeUnavailable[len(a.LastTimeUnavailable)-1].Format("2006-01-02 15:04:05"),
			time.Since(a.LastTimeUnavailable[len(a.LastTimeUnavailable)-1]).Round((time.Millisecond))))
	} else if len(a.LastTimeAvailable) > len(a.LastTimeUnavailable) {
		res.WriteString(fmt.Sprintf(green("STATUS: UP, Availability: %v, Since: %v, Duration: %v\n"), a.Availability, a.LastTimeAvailable[len(a.LastTimeAvailable)-1].Format("2006-01-02 15:04:05"),
			time.Since(a.LastTimeAvailable[len(a.LastTimeAvailable)-1]).Round(time.Millisecond)))
	}
	res.WriteString(fmt.Sprintf("Lower Threshold Violation	|	Upper Thereshold Violation\n"))
	for i := 0; i < len(a.LastTimeUnavailable); i++ {
		res.WriteString(fmt.Sprint(a.LastTimeUnavailable[i].Format("2006-01-02 15:04:05"), "   "))
		if i+1 < len(a.LastTimeAvailable) {
			res.WriteString(fmt.Sprintf("%v\n", a.LastTimeAvailable[i].Format("2006-01-02 15:04:05")))
		}
	}
	return res.String()
}

// Function that prints the alert
func (a *Alert) Print() {
	red := color.FgRed.Render
	green := color.FgGreen.Render
	if len(a.LastTimeAvailable) == len(a.LastTimeUnavailable) {
		fmt.Printf(red("STATUS: DOWN, Availability: %v, Since: %v, Duration: %v\n"), a.Availability, a.LastTimeUnavailable[len(a.LastTimeUnavailable)-1].Format("2006-01-02 15:04:05"),
			time.Since(a.LastTimeUnavailable[len(a.LastTimeUnavailable)-1]))
	} else if len(a.LastTimeAvailable) > len(a.LastTimeUnavailable) {

		fmt.Printf(green("STATUS: UP, Availability: %v%%, Since: %v, Duration: %v\n"), a.Availability, a.LastTimeAvailable[len(a.LastTimeAvailable)-1].Format("2006-01-02 15:04:05"),
			time.Since(a.LastTimeAvailable[len(a.LastTimeAvailable)-1]))
	}
	fmt.Printf("Lower Threshold Violation	|	Upper Thereshold Violation\n")
	for i := 0; i < len(a.LastTimeUnavailable); i++ {
		fmt.Print(a.LastTimeUnavailable[i])
		if i+1 < len(a.LastTimeAvailable) {
			fmt.Print(a.LastTimeUnavailable[i].Format("2006-01-02 15:04:05"), "   ")
		}
		fmt.Println()
	}
}
