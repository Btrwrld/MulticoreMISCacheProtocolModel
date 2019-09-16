package Core

import (
	"../GUI"
	"fmt"
	"strconv"
	"time"
)

type Core struct{
	Id             	int
	CPU            	CPU
	CacheController	CacheController
	storeDirectory 	string
	gui				GUI.CoreGUI
	clock			*chan bool
	missChan		chan bool
}

func (core *Core) Init(id int, memory *Memory, clock *chan bool, gui GUI.CoreGUI) {
	// Sets every component of the CPU
	core.Id = id
	core.CPU = CPU{}
	core.CacheController = CacheController{}
	core.clock = clock
	core.storeDirectory = "/home/erick/googleDrive/TEC/ArquiII/Proyecto1/Output/Core" + strconv.Itoa(id) + ".json"

	// Initializes to defaults
	core.CPU.Init(id, len(memory.Block), clock)
	core.missChan = make(chan bool)
	core.CacheController.Init(memory.bus, id, gui, &core.missChan)
	core.gui = gui
}

func (core *Core) Run() {
	// Create chanel to check wen the instruction execution is finished
	done := make(chan bool)
	ok := true

	// Generate initial state
	core.ShowState()

	// Activate cache snooping
	go core.CacheController.snoop()

	// Start executing instructions
	go core.CPU.executeInstruction(&core.CacheController, &done)


	// When finish an instruction save state
	for ok{

		// Wait for the message
		select {
		// Either show state
			case _, ok =<- done:
				// show state
				core.ShowState()

			// Or report miss
			case rw:=<- core.missChan:
				core.gui.Mutex.Lock()
				if rw{
					core.gui.Miss.SetText("Read Miss!")
				} else{
					core.gui.Miss.SetText("Write Miss!")
				}

				time.Sleep(time.Second * 2)
				core.gui.Miss.SetText("")
				core.gui.Mutex.Unlock()
		}




	}
}


func (core *Core) Print() {
	fmt.Println("-------------------------------------------")
	fmt.Println("Core 		", 						core.Id)
	fmt.Println("CPU info: 	", 					core.CPU)
	fmt.Println("CacheController info: 	", 		core.CacheController)
	fmt.Println("-------------------------------------------")
}

func (core *Core) ShowState(){

	// Show cache value
	core.CacheController.mutex.Lock()
	core.gui.Mutex.Lock()
	var s string
	for i := 0; i < 8; i++ {

		switch core.CacheController.Block[i].State {

		case INVALID:
			s = "I"

		case MODIFIED:
			s = "M"

		case SHARED:
			s = "S"

		}
		// Add tag
		s += strconv.Itoa(core.CacheController.Block[i].Tag)

		core.gui.Cache[i].SetText(strconv.Itoa(core.CacheController.Block[i].Data) + "\t" + s)
	}
	core.gui.Mutex.Unlock()
	core.CacheController.mutex.Unlock()

	//Show instructions
	var toShow string
	toShow += "\tNext:"
	for i := 0; i < len(core.CPU.Instructions); i++ {
		inst := core.CPU.Instructions[i]

		//If it hasnt been solved
		if !inst.IsFinished{
			switch inst.InstructionType {
			case PROCESS:
				toShow += "\tProcess\n"
			case READ:
				toShow += "\tRead at " + strconv.Itoa(inst.TargetBlockId) +"\n"
			case WRITE:
				toShow += "\tWrite at " + strconv.Itoa(inst.TargetBlockId) +"\n"
			}
		}
	}

	fmt.Println("Core ", core.Id, " updated")
	core.gui.Mutex.Lock()
	core.gui.Inst.SetText(toShow)
	core.gui.Mutex.Unlock()
}