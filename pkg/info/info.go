package info

import (
	"fmt"
	"strings"
	"time"

	"github.com/iwita/monitoring/alert"
	"github.com/iwita/monitoring/heap"
)

type Result struct {
	Max          time.Duration
	Average      time.Duration
	Percentile   time.Duration
	Availability float64
	StatusCodes  string
}

type Response struct {
	Delay  time.Duration
	Status int
}

// Info is the main type of this package.
// It inlcudes both raw and processed information about a specific time window
// specified by 'Duration'
// *Additionally, some of the Info types may include an Alert
type Info struct {
	MaxResponse         time.Duration
	AverageResponse     time.Duration
	SumResponses        time.Duration
	ResponsesList       []*Response
	MaxResponsesList    []time.Duration
	Length              int
	Duration            time.Duration
	StatusCodesCount    map[int]int
	SuccessfulResponses int
	TotalResponses      int
	hasAlert            bool
	Alert               *alert.Alert
}

func NewInfo(duration, interval time.Duration, hasAlert bool) *Info {
	var length int
	if duration == time.Duration(0) {
		// we have the unlimited case
		length = -1
	} else {
		length = int(duration / interval)
	}
	i := &Info{
		MaxResponse:         0,
		ResponsesList:       make([]*Response, 0),
		MaxResponsesList:    make([]time.Duration, 0),
		SumResponses:        time.Duration(0) * time.Millisecond,
		Duration:            duration,
		Length:              length,
		StatusCodesCount:    make(map[int]int, 0),
		SuccessfulResponses: 0,
		TotalResponses:      0,
		hasAlert:            false,
	}
	if hasAlert {
		i.Alert = alert.NewAlert(0.8)
		i.hasAlert = true
	}
	return i
}

// Updates the information stored in a predefined time window
func (i *Info) Update(status int, elapsedTime time.Duration) {
	// 1. Delete the outdated responses if any
	if i.TotalResponses == i.Length {
		i.TotalResponses--
		responseToBeDeleted := i.ResponsesList[0]
		i.ResponsesList = i.ResponsesList[1:]
		i.SumResponses -= responseToBeDeleted.Delay
		i.StatusCodesCount[responseToBeDeleted.Status]--
		if responseToBeDeleted.Status >= 200 && responseToBeDeleted.Status < 300 {
			i.SuccessfulResponses--
		}
		// Update the maximum in the respective Deque
		if responseToBeDeleted.Delay == i.MaxResponsesList[0] {
			i.MaxResponsesList = i.MaxResponsesList[1:]
		}
	}

	// 2. Push a new item

	// 2.1 Update the maximum in the helping data structure
	if i.TotalResponses == 0 {
		i.MaxResponsesList = append(i.MaxResponsesList, elapsedTime)
	} else {
		// Update the max in the helping data structure
		for j, el := range i.MaxResponsesList {
			// remove all elements smaller than current
			// This helping array stores Max elements in the order they occur
			// This way, if  the current max elements needs to be deleted and be excluded from the time window,
			// we have the next max value stored
			if el < elapsedTime {
				//i.MaxResponsesList[j] = elapsedTime
				i.MaxResponsesList = i.MaxResponsesList[:j]
				break
			}
		}
		i.MaxResponsesList = append(i.MaxResponsesList, elapsedTime)
	}

	// 2.2 Add info about the new item

	i.StatusCodesCount[status]++
	if status >= 200 && status < 300 {
		i.SuccessfulResponses++
		// Keep the sum of the delays in the time window, in order to
		// calculate the average in constant time
		i.SumResponses += elapsedTime

	}
	i.TotalResponses++
	i.ResponsesList = append(i.ResponsesList, &Response{
		Delay:  elapsedTime,
		Status: status,
	})

	// Moved upwards only in case of successful response
	//i.SumResponses += elapsedTime

	if i.hasAlert {
		i.UpdateAlert()
	}
}

// Updates the alert's values
// More specifically,
// 1. Stores the current availability in the array
// 2. If the current state is available, and needs to change, it stores the current time
//    and moves to unavailable state.
//    Else if the current state is unavailable, and needs to change, it stores the current time
//    and moves back to the available state.
func (i *Info) UpdateAlert() {
	i.Alert.Availability = float64(float64(i.SuccessfulResponses) / float64(i.TotalResponses))
	switch i.Alert.AlertState {
	case alert.Available:
		if i.Alert.Availability < i.Alert.Threshold {
			i.Alert.LastTimeUnavailable = append(i.Alert.LastTimeUnavailable, time.Now())
			i.Alert.AlertState++
		}
	case alert.Unavailable:
		if i.Alert.Availability >= i.Alert.Threshold {
			i.Alert.LastTimeAvailable = append(i.Alert.LastTimeAvailable, time.Now())
			i.Alert.AlertState--
		}
	}
}

// Prints the information stored
func (i *Info) PrintInfo() {
	if i.TotalResponses == 0 {
		fmt.Println("Metrics currently unavailable")
		return
	}
	// Calculate the 90th percentile of the responses time
	percentile := get90thPercentile(i.ResponsesList)

	average := time.Duration(int(i.SumResponses) / i.TotalResponses)
	max := i.MaxResponsesList[0]
	fmt.Printf("(Average/Max/90th percentile) response time: (%v/%v/%v)\n", average, max, percentile)
	for key, val := range i.StatusCodesCount {
		fmt.Printf("Status %v => %v\n", key, val)
	}
	fmt.Printf("Availability: %v%% \n", i.SuccessfulResponses*100/i.TotalResponses)
}

func (i *Info) GetResult() *Result {
	result := &Result{}
	if i.TotalResponses == 0 {
		return nil
	}

	// Calculate the 90th percentile of the responses time
	result.Percentile = get90thPercentile(i.ResponsesList).Round(time.Millisecond)
	result.Average = time.Duration(int(i.SumResponses) / i.TotalResponses).Round(time.Millisecond)
	result.Max = i.MaxResponsesList[0].Round(time.Millisecond)

	temp := strings.Builder{}
	for key, val := range i.StatusCodesCount {
		fmt.Fprintf(&temp, "status %v => %v\n", key, val)
	}
	result.StatusCodes = temp.String()
	result.Availability = (float64(i.SuccessfulResponses) * 100 / float64(i.TotalResponses))
	return result
}

func get90thPercentile(responses []*Response) time.Duration {
	size := 0.1 * float64(len(responses))
	minHeap := heap.NewMinHeap(int(size))
	j := 0
	for j < int(size) {
		minHeap.Insert(int(responses[j].Delay))
		j++
	}
	if minHeap.Size > 0 {
		for j < len(responses) && minHeap.Size > 0 {
			if int(responses[j].Delay) > minHeap.Peek() {
				minHeap.Remove()
				minHeap.Insert(int(responses[j].Delay))
			}
			j++
		}
		return time.Duration(minHeap.Peek())
	}
	return -1
}
