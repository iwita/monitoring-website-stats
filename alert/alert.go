package alert

type State uint32

const (

	// Initial dummy state
	Unknown State = iota

	// Normal State -> Available > 80% Availability
	Available

	//Transition to Unavailable
	FromAvailableToUnavailable

	// Unavailable -> Over the last 2 minutes, availability is below 80%
	Unavailable

	// Transition to Available
	FromUnavailableToAvailable
)

type Alert struct {
	AlertState State
	Threshold  float64
}

func NewAlert(t float64) *Alert {
	return &Alert{
		AlertState: Unknown,
		Threshold:  t,
	}
}

func (a *Alert) Exec() {

	for {
		switch a.AlertState {
		case Unknown:
			// print no data yet
		case Available:
			// check Info (2m)
			// if need to change state -> goto TransitionToUnavailable
		case FromAvailableToUnavailable:
			// print that last 2 minutes website was unavailable
			// -> goto Unavailable state
		case Unavailable:
			// check if the state needs to be changed
			// print -> goto
		case FromUnavailableToAvailable:
			// print the transition
			// goto -> Available state
		}
	}
}
