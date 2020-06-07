package info

import (
	"time"
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
}

func NewInfo(duration, interval time.Duration) *Info {
	var length int
	if duration == time.Duration(0) {
		// we have the unlimited case
		length = -1
	} else {
		length = int(duration / interval)
	}
	return &Info{
		MaxResponse:         0,
		ResponsesList:       make([]*Response, 0),
		MaxResponsesList:    make([]time.Duration, 0),
		SumResponses:        time.Duration(0) * time.Millisecond,
		Duration:            duration,
		Length:              length,
		StatusCodesCount:    make(map[int]int, 0),
		SuccessfulResponses: 0,
		TotalResponses:      0,
	}
}

func (i *Info) Update(status int, elapsedTime time.Duration) {

	// Delete the outdated responses if any
	if i.TotalResponses == i.Length {
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
	}

	// Add info about the new item

	// Update the max in the helping data structure
	//tmp := make([]time.Duration, 0)
	for j, el := range i.MaxResponsesList {
		// remove all elements smaller than current
		if el < elapsedTime {
			i.MaxResponsesList[j] = elapsedTime
			i.MaxResponsesList = i.MaxResponsesList[:j+1]
		}

	}
	// add current at the end of the deque

	i.StatusCodesCount[status]++
	if status == 200 {
		i.SuccessfulResponses++
		i.TotalResponses++
	}
	i.ResponsesList = append(i.ResponsesList, &Response{
		Delay:  elapsedTime,
		Status: status,
	})
	if elapsedTime > i.MaxResponse {
		i.MaxResponse = elapsedTime
	}
	i.SumResponses += elapsedTime

	// Add the newly added one

	i.ResponsesList = append(i.ResponsesList, &Response{
		Delay:  elapsedTime,
		Status: status,
	})

}
