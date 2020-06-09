package alert

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestUP(t *testing.T) {
	want := strings.Builder{}
	a := NewAlert(0.8)
	start := time.Now()
	a.Availability = 90
	res := a.PrintTest()
	want.WriteString(fmt.Sprintf("STATUS: UP, Availability: 90, Since: %v\n", start.Format("2006-01-02 15:04:05")))
	want.WriteString("Lower Threshold Violation	|	Upper Thereshold Violation\n")
	if res != want.String() {
		t.Errorf("Got = %s; want %s", res, want.String())
	} else {
		fmt.Println("Test1 OK")
	}
	// a.Availability = 0.78
	// time.Sleep(3 * time.Minute)
	// a.Availability = 0.81
	// time.Sleep()
}

func TestUpDownTransition(t *testing.T) {
	want := strings.Builder{}
	a := NewAlert(0.8)
	a.Availability = 70

	// Goes down
	a.AlertState++

	downAt := time.Now()
	a.LastTimeUnavailable = append(a.LastTimeUnavailable, downAt)
	res := a.PrintTest()

	want.WriteString(fmt.Sprintf("STATUS: DOWN, Availability: %v, Since: %v\n", a.Availability, downAt.Format("2006-01-02 15:04:05")))
	want.WriteString(fmt.Sprintf("Lower Threshold Violation	|	Upper Thereshold Violation\n"))
	want.WriteString(fmt.Sprint(downAt.Format("2006-01-02 15:04:05"), "   "))

	if res != want.String() {
		t.Errorf("Got = %s; want %s", res, want.String())
	} else {
		fmt.Println("Test2 OK")
	}
}

func TestUpDownUpTransition(t *testing.T) {
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

	want.WriteString(fmt.Sprintf("STATUS: UP, Availability: 90, Since: %v\n", upAt.Format("2006-01-02 15:04:05")))

	want.WriteString(fmt.Sprintf("Lower Threshold Violation	|	Upper Thereshold Violation\n"))
	want.WriteString(fmt.Sprint(downAt.Format("2006-01-02 15:04:05"), "   "))
	want.WriteString(fmt.Sprintf("%v\n", upAt.Format("2006-01-02 15:04:05")))

	if res != want.String() {
		t.Errorf("Got = %s; want %s", res, want.String())
	} else {
		fmt.Println("Test3 OK")
	}
}
