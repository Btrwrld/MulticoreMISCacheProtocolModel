package Core

type Bus struct {
	// Memory serial mutex, so each cc can communicate with memory
	EventBus  [5]chan *Event      // Each core has a bus and one for memory
	AnswerBus [4]chan MemoryBlock // Each core recives answers here
}

func (bus *Bus) Init(){
	// Create buses
	for i := 0; i < 4; i++ {
		bus.EventBus[i] = make(chan *Event, 4)
		bus.AnswerBus[i] = make(chan MemoryBlock)
	}
	// This is the memory bus
	bus.EventBus[4] = make(chan *Event, 4)

}

func (bus *Bus) Broadcast(senderId int, e *Event){

	// Send the event to every one but the sender
	for i := 0; i < 5; i++ {
		if i != senderId{
			bus.EventBus[i] <- e
		}
	}
}








