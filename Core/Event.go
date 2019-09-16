package Core

import (
	"strconv"
	"sync"
)

// Things each cc can do
type Action int
const (  // iota is reset to 0
	READING 	Action = iota // READING = 0
	WRITING  	Action = iota // WRITING = 1
	INVALIDATE	Action = iota // INVALIDATE = 2
)

// Result of the action
type Result int
const (  // iota is reset to 0
	MISS 	Result = iota // MISS = 0
	HIT  	Result = iota // HIT = 1
)

type Event struct {
	// Define the characteristics of the given event
	Action            Action     // The action we are executing
	Result            Result     // The result of the attempted action
	TargetMemoryBlock MemoryBlock// Info of the cache block used
	CoreId            int        // The id of the core executing the action
	isResolved        bool       // If someone already took care of the problem: true, else: false
	mutex             sync.Mutex // Used when solving the issue
}

// Checks if the event has been solved, if it hasn't been solved
// IT SETS TO SOLVED so the one who asks when the event hast been solved
// has the responsibility to solve it.
func (event *Event) isSolved() bool{

	// Lock the function and unlock it when the return is done
	event.mutex.Lock()
	defer event.mutex.Unlock()

	// If the event is not solved then solve it and say it wasn't solved
	if !event.isResolved {
		event.isResolved = true
		return false

	} else{
		// If the event has been solved return true
		return true
	}

}


func (event *Event) getInfo() string{

	// Lock the function and unlock it when the return is done
	event.mutex.Lock()
	defer event.mutex.Unlock()

	var e string

	switch event.Action {
	case READING:
		e += "Read "
	case WRITING:
		e += "Write "
	case INVALIDATE:
		e += "Invalidate "
	}

	switch event.Result {
	case HIT:
		e += " hit in "
	case MISS:
		e += " miss in "
	}

	return e + strconv.Itoa(event.TargetMemoryBlock.Id) + " from Core" + strconv.Itoa(event.CoreId)

}
