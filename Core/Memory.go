package Core

import (
	"../GUI"
	"strconv"
	"sync"
	"time"
)

const numMemoryBlocks = 16

// Define the blocks the build the memory
type MemoryBlock struct {
	Id   int
	Data int
}

// Define a memory with n blocks
type Memory struct {
	bus            *Bus
	writeTime      time.Duration
	readTime       time.Duration
	storeDirectory string
	gui            GUI.MemoryGUI
	// Mutex to prevent reads and writes at the same time
	mutex sync.Mutex
	Block [numMemoryBlocks]MemoryBlock
}

func (memory *Memory) Init(bus *Bus, gui GUI.MemoryGUI) {

	// Initialize memory to its default values
	for i := 0; i < numMemoryBlocks; i++ {
		memory.Block[i] = MemoryBlock{i, 0}
	}

	memory.bus = bus
	memory.writeTime = time.Second * 5
	memory.readTime = time.Second * 3
	memory.storeDirectory = "/home/erick/googleDrive/TEC/ArquiII/Proyecto1/Output/MemoryState.json"
	memory.gui = gui
}

func (memory *Memory) ReadBlock(blockId int) int {

	// Place a mutex to lock the memory
	memory.mutex.Lock()

	// Return the data and unlock the memory
	defer memory.mutex.Unlock()
	return memory.Block[blockId].Data

}

func (memory *Memory) WriteBlock(blockId int, data int) {

	// Place a mutex to lock the memory
	memory.mutex.Lock()

	// Set the value
	memory.Block[blockId].Data = data

	// Unlock the memory
	memory.mutex.Unlock()

}

func (memory *Memory) Run(clock *chan bool) {

	// Store initial state
	memory.showState()

	for {

		select {
		case <-*clock:
			// This message is sent just before the tok
			// so everyone should have ended their processes
			go memory.showState()

			// Wait for a message to arrive
		case e := <-memory.bus.EventBus[4]:
			// Solve the message in another thread so we dont miss any incoming msg
			go memory.solveEvent(e)
		}
	}
}

func (memory *Memory) solveEvent(e *Event) {

	isValidEvent := e.Action != INVALIDATE

	//Penalize memory access
	if e.Action == READING {
		time.Sleep(memory.readTime)
	} else if e.Action == WRITING {
		time.Sleep(memory.writeTime)
	}

	println(e.getInfo() + " into memory")
	// If its a valid event and it hasn't been solved
	if isValidEvent && !e.isSolved() {

		// If its a write hit, write the desired block
		if e.Action == WRITING && e.Result == HIT {
			memory.WriteBlock(e.TargetMemoryBlock.Id, e.TargetMemoryBlock.Data)

			// If its a read miss try to return the value
		} else if e.Action == READING && e.Result == MISS {
			// Get data from the block
			r := MemoryBlock{e.TargetMemoryBlock.Id, memory.ReadBlock(e.TargetMemoryBlock.Id)}
			// Send the block
			memory.bus.AnswerBus[e.CoreId] <- r

		}
	}
}

func (memory *Memory) showState() {

	memory.mutex.Lock()
	memory.gui.Mutex.Lock()
	for i := 0; i < len(memory.Block); i++ {
		memory.gui.Line[i].SetText(strconv.Itoa(memory.Block[i].Data))
	}
	memory.gui.Mutex.Unlock()
	memory.mutex.Unlock()
}
