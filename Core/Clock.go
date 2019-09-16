package Core

import (
	"time"
)

type Clock struct{
	Clock  		chan bool
	MemClock	chan bool
	UserControl	*chan bool
	Next		*chan bool
	period 		time.Duration
}

func (clock *Clock) Init(numCores int, period time.Duration){
	// Initializes the clock default value
	clock.period = period
	clock.Clock = make(chan bool, numCores) // Set buffer for each core
	clock.MemClock = make(chan bool)

}

func (clock *Clock) Start(isUserControlled bool){


	for {

		// Tell the memory to store the state
		clock.MemClock <- true

		if isUserControlled{
			*clock.Next <- false
			_, ok :=<- *clock.UserControl
			*clock.Next <- true
			// If the channel is closed, we are over
			if !ok {
				break
			}
		}

		// Send a tick for each core
		for i := 0; i < cap(clock.Clock); i++ {
			clock.Clock <- true
		}

		// Maintain the state for some time
		time.Sleep(clock.period)


		// Send a tok for each core
		for i := 0; i < cap(clock.Clock); i++ {
			clock.Clock <- false
		}

		// Maintain the state for some time
		time.Sleep(clock.period)

	}

	// Close the channel so the cores know its over
	close(clock.Clock)


}