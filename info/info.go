package info

import (
	"fmt"
	"time"

	"github.com/iwita/monitoring/alert"
)

type Response struct {
	Delay  time.Duration
	Status int
}

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

func (i *Info) Update(status int, elapsedTime time.Duration) {

	// Delete the outdated responses if any
	if i.TotalResponses == i.Length {
		i.TotalResponses--
		responseToBeDeleted := i.ResponsesList[0]
		i.SumResponses -= responseToBeDeleted.Delay
		i.StatusCodesCount[responseToBeDeleted.Status]--
		if responseToBeDeleted.Status == 200 {
			i.SuccessfulResponses--
		}
		// Update the maximum in the respective Deque
		if responseToBeDeleted.Delay == i.MaxResponsesList[0] {
			i.MaxResponsesList = i.MaxResponsesList[1:]
		}
	}

	// Push a new item

	// Update the maximum in the helping data structure
	if i.TotalResponses == 0 {
		i.MaxResponsesList = append(i.MaxResponsesList, elapsedTime)
	} else {
		// Update the max in the helping data structure
		for j, el := range i.MaxResponsesList {
			// remove all elements smaller than current
			if el < elapsedTime {
				i.MaxResponsesList[j] = elapsedTime
				i.MaxResponsesList = i.MaxResponsesList[:j+1]
			}
		}
	}

	// Add info about the new item

	i.StatusCodesCount[status]++
	if status >= 200 && status <= 300 {
		i.SuccessfulResponses++
	}
	i.TotalResponses++
	i.ResponsesList = append(i.ResponsesList, &Response{
		Delay:  elapsedTime,
		Status: status,
	})
	i.SumResponses += elapsedTime

	if i.hasAlert {
		i.UpdateAlert()
	}
}

func (i *Info) UpdateAlert() {
	i.Alert.Availability = float64(float64(i.SuccessfulResponses) * 100 / float64(i.TotalResponses))
	switch i.Alert.AlertState {
	//case alert.Unknown:
	case alert.Available:
		if i.Alert.Availability < i.Alert.Threshold {
			i.Alert.LastTimeUnavailable = append(i.Alert.LastTimeUnavailable, time.Now())
			i.Alert.AlertState++
		}
	//case alert.FromAvailableToUnavailable:
	//case alert.FromUnavailableToAvailable:
	case alert.Unavailable:
		if i.Alert.Availability > i.Alert.Threshold {
			i.Alert.LastTimeUnavailable = append(i.Alert.LastTimeAvailable, time.Now())
			i.Alert.AlertState--
		}
	}
}

func (i *Info) PrintInfo() {
	if i.TotalResponses == 0 {
		fmt.Println("Metrics currently unavailable")
		return
	}
	average := time.Duration(int(i.SumResponses) / i.TotalResponses)
	max := i.MaxResponsesList[0]
	fmt.Printf("Average/Max response time: %v/%v\n", average, max)
	for key, val := range i.StatusCodesCount {
		fmt.Printf("Status %v => %v\n", key, val)
	}
	//fmt.Println()
	fmt.Printf("Availability: %v%% \n", i.SuccessfulResponses*100/i.TotalResponses)
	fmt.Println()
}
