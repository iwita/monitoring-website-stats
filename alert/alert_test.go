package alert

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gookit/color"
)

// Test the availability while it is greater than the threshold(80)
func TestUP(t *testing.T) {
	green := color.FgGreen.Render

	want := strings.Builder{}
	a := NewAlert(0.8)
	start := time.Now()
	a.Availability = 90
	res := a.PrintTest()
	want.WriteString(fmt.Sprintf(green("STATUS: UP, Availability: 90, Since: %v, Duration: %v\n"), start.Format("2006-01-02 15:04:05"),
		time.Since(start).Round(time.Millisecond)))
	want.WriteString("Unavailable		|	Available again\n")
	if res != want.String() {
		t.Errorf("Got: %s\nWant: %s", res, want.String())
	} else {
		fmt.Println("Up OK")
	}
}

// Test the transition of the availability from >80 to <80
func TestUpDownTransition(t *testing.T) {
	red := color.FgRed.Render
	want := strings.Builder{}
	a := NewAlert(0.8)
	a.Availability = 70

	// Goes down
	a.AlertState++

	downAt := time.Now()
	a.LastTimeUnavailable = append(a.LastTimeUnavailable, downAt)
	res := a.PrintTest()

	want.WriteString(fmt.Sprintf(red("STATUS: DOWN, Availability: %v, Since: %v, Duration: %v\n"), a.Availability, downAt.Format("2006-01-02 15:04:05"),
		time.Since(downAt).Round(time.Millisecond)))
	want.WriteString("Unavailable		|	Available again\n")
	want.WriteString(fmt.Sprint(downAt.Format("2006-01-02 15:04:05"), "		"))

	if res != want.String() {
		t.Errorf("Got: %s\nWant: %s", res, want.String())
	} else {
		fmt.Println("Up to Down - OK")
	}
}

// Test the transition from < 80 to > 80
func TestUpDownUpTransition(t *testing.T) {
	green := color.FgGreen.Render
	want := strings.Builder{}
	a := NewAlert(0.8)
	a.Availability = 70

	// Goes down
	a.AlertState++

	downAt := time.Now()
	a.LastTimeUnavailable = append(a.LastTimeUnavailable, downAt)

	// Goes up again
	upAt := time.Now()
	a.AlertState--
	a.LastTimeAvailable = append(a.LastTimeAvailable, upAt)
	a.Availability = 90
	res := a.PrintTest()

	want.WriteString(fmt.Sprintf(green("STATUS: UP, Availability: 90, Since: %v, Duration: %v\n"), upAt.Format("2006-01-02 15:04:05"), time.Since(upAt).Round(time.Millisecond)))

	want.WriteString("Unavailable		|	Available again\n")
	want.WriteString(fmt.Sprint(downAt.Format("2006-01-02 15:04:05"), "		"))
	want.WriteString(fmt.Sprintf("%v\n", upAt.Format("2006-01-02 15:04:05")))

	if res != want.String() {
		t.Errorf("Got: %s\nWant: %s", res, want.String())
	} else {
		fmt.Println("Down to Up - OK")
	}
}
