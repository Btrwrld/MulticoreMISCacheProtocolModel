package Core

import (
	"../GUI"
	"fmt"
	"strconv"
	"sync"
)

const numCacheBlocks = 8

// CacheController Block states
type State int
const (  // iota is reset to 0
	MODIFIED 	State = iota // PRIVATE = 0
	SHARED  	State = iota // SHARED = 1
	INVALID 	State = iota // INVALID = 2
)

// CacheController Block structure
type CacheBlock struct{
	Id      int
	Tag     int
	Data    int
	State   State
}

// CacheController built of n blocks
type CacheController struct{
	Id            	int
	bus				*Bus
	gui				GUI.CoreGUI
	missChan		*chan bool
	mutex 			*sync.Mutex
	Block         	[numCacheBlocks]CacheBlock

}

func (cache *CacheController) Init(bus *Bus, id int, gui GUI.CoreGUI, missChan *chan bool) {

	// Initialize CacheController to its default values
	for i := 0; i < numCacheBlocks ; i++  {
		cache.Block[i] = CacheBlock{i, i, 0, INVALID}
	}
	cache.Id = id
	cache.bus = bus
	cache.gui = gui
	cache.mutex = &sync.Mutex{}
	cache.missChan = missChan
}

func (cache *CacheController) ReadBlock(memoryBlockId int) int {

	// Lock read and write memory access
	cache.mutex.Lock()

	// Calc the cache id of the memory block
	cacheId := (memoryBlockId ) % (len(cache.Block))
	// Create the message variable
	var message MemoryBlock

	// Read Miss
	// Checks if the cache block represents the desired block or the block is invalid
	if cache.Block[cacheId].Tag != memoryBlockId || cache.Block[cacheId].State == INVALID {
		// Say there was a miss and penalize
		fmt.Println("Read miss at Core", cache.Id, " in block ", memoryBlockId )
		*cache.missChan <- true

		// If the cache state is in modified but it isn't the block we want
		// then we should write back to memory before erasing it
		if cache.Block[cacheId].State == MODIFIED {

			// Build the message
			message = MemoryBlock{cache.Block[cacheId].Tag, 0}

			// Send a write hit so memory updates and other caches get invalidated
			eMod := &Event{WRITING, HIT, message, cache.Id, false, sync.Mutex{}}
			// Tell everybody
			cache.bus.Broadcast(cache.Id,eMod)
		}

		// Build message
		message = MemoryBlock{memoryBlockId, 0}

		// Now we can safely report the miss
		// Create the event describing the problem
		e := &Event{READING, MISS, message, cache.Id,false, sync.Mutex{}}
		// Send it to every one
		cache.bus.Broadcast(cache.Id, e)

		// Wait for the response from some one, since reading from memory takes longer,
		// if any cache sees it it will answer first
		realValue :=<- cache.bus.AnswerBus[cache.Id]

		// Update the cache block
		cache.Block[cacheId].Data = realValue.Data
		cache.Block[cacheId].Tag = realValue.Id
		// The state is shared because we are  just reading
		cache.Block[cacheId].State = SHARED
		// Now we can return the value
	}

	// Unlock mutex after return
	defer cache.mutex.Unlock()

	// Data in the cache block checked is correct
	return cache.Block[cacheId].Data



}

func (cache *CacheController) WriteBlock(memoryBlockId int, data int) {

	// Lock read and write memory access
	cache.mutex.Lock()

	// Calc the cache id of the memory block
	cacheId := (memoryBlockId ) % (len(cache.Block))
	// Create the message
	var message MemoryBlock

	// Write Miss
	// Checks if the cache block represents the desired block or the block is invalid
	if cache.Block[cacheId].Tag != memoryBlockId || cache.Block[cacheId].State == INVALID{
		// Say there was a miss and penalize
		fmt.Println("Write miss at Core" + strconv.Itoa(cache.Id))
		*cache.missChan <- false

		// If the cache state is in modified but it isn't the block we want
		// then we should write back to memory before erasing it
		if cache.Block[cacheId].State == MODIFIED{

			// Build the message
			message = MemoryBlock{cache.Block[cacheId].Tag, cache.Block[cacheId].Data}

			// Send a write hit so memory updates and other caches get invalidated
			eMod := &Event{WRITING, HIT, message, cache.Id,false, sync.Mutex{}}
			// Tell everybody
			cache.bus.Broadcast(cache.Id, eMod)
		}

		// Build the miss message
		message = MemoryBlock{memoryBlockId, data}

		// Now we can safely report the miss
		// Create the event describing the problem
		e := &Event{WRITING, MISS, message, cache.Id,false, sync.Mutex{}}
		// Send it to every one
		cache.bus.Broadcast(cache.Id, e)


	// If we are on SHARED we should place an invalidate on bus
	}else if cache.Block[cacheId].State == SHARED{

		// Build invalidate for the other blocks
		message = MemoryBlock{memoryBlockId, data}

		// Create the invalidate event on bus
		eShared := &Event{INVALIDATE, HIT, message, cache.Id,false, sync.Mutex{}}
		// Send the event
		cache.bus.Broadcast(cache.Id, eShared)
	}

	// Finally after placing the event in case of shared or fetching the block in case of miss
	// or just being in MODIFIED, update the block
	cache.Block[cacheId].Data = data
	cache.Block[cacheId].Tag = memoryBlockId
	cache.Block[cacheId].State = MODIFIED


	// Unlock mutex
	cache.mutex.Unlock()

}


func (cache *CacheController) snoop(){

	var e *Event
	id := cache.Id

	for{
		// Wait for a message to arrive listening your personal bus
		e =<- cache.bus.EventBus[id]
		//Resolve it
		go cache.solveEvent(e)

	}

}

func (cache *CacheController) solveEvent(e *Event){

	cache.mutex.Lock()
	// Define the cache block realated with the event (if we have it)
	// And the mesagge variable for the broadcasts
	var cacheBlock *CacheBlock
	var message MemoryBlock
	isInCache := cache.Block[(e.TargetMemoryBlock.Id) % len(cache.Block)].Tag == e.TargetMemoryBlock.Id
	if isInCache{
		cacheBlock = &cache.Block[(e.TargetMemoryBlock.Id ) % len(cache.Block)]
	}

	// Check we got the message
	println(e.getInfo() + " into cc" + strconv.Itoa(cache.Id))
	// If the event was sent from a core that isn't ours
	// or affects a block that we dont have
	if e.CoreId != cache.Id && isInCache {

		switch e.Action {

		case INVALIDATE:
			// If the cache we have is the one we need to invalidate is in shared
			if cacheBlock.State == SHARED{
				// Invalidate the block
				cacheBlock.State = INVALID
			}

		case WRITING:
			// If we are writting a block that is M somewhere else
			if e.Result == MISS{
				if cacheBlock.State == MODIFIED {
					// Build the message
					message = MemoryBlock{cacheBlock.Tag, cacheBlock.Data}

					// Send a write hit so memory updates and other caches get invalidated
					eMod := &Event{WRITING, HIT, message, cache.Id,false, sync.Mutex{}}
					// Tell everybody
					cache.bus.Broadcast(cache.Id, eMod)
				}
				// If we get a write miss in a shared block invalidate
				// Invalidate the block
				cacheBlock.State = INVALID

			}

		case READING:


			if e.Result == MISS{

				if cacheBlock.State != INVALID{
					// If someone is reading a block that is M here, write back and share it
					if cacheBlock.State == MODIFIED {
						// Build the message
						message = MemoryBlock{cacheBlock.Tag, cacheBlock.Data}

						// Send a write hit so memory updates and other caches get invalidated
						eMod := &Event{WRITING, HIT, message, cache.Id,false, sync.Mutex{}}
						// Tell everybody
						cache.bus.Broadcast(cache.Id,eMod)
						// Share the block
						cacheBlock.State = SHARED

					}

					// Else stay in shared

					// Share data only is it hasn't been solved by someone else
					if !e.isSolved(){
						cache.bus.AnswerBus[e.CoreId] <- MemoryBlock{cacheBlock.Tag, cacheBlock.Data}
						fmt.Println("Sharing data from ", cache.Id, " to ", e.CoreId)
					}

				}

			}

		
		}

	}


	cache.mutex.Unlock()

}
