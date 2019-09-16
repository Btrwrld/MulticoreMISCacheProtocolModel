package Core

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"time"
)

const instructionPath  = "/home/erick/googleDrive/TEC/ArquiII/Proyecto1/Output/iCPU"
const maxInstructions = 20

type CPU struct{
	Id                 int
	Instructions       [maxInstructions]Instruction
	clock				*chan bool
}

func (cpu *CPU) Init(id int, numMemoryBlocks int, clock *chan bool) {
	cpuid := strconv.Itoa(id) + ".csv"
	fmt.Println("Generating Instructions for CPU", cpuid, "...")

	// Build the output csv file with the Instructions generated
	var output bytes.Buffer
	output.WriteString("Instruction, Instruction number, Target memory Block\n")
	for i := 0; i < maxInstructions; i++ {
		// Generate Instructions from a normal distribution with std=1 and mean=1
		iType := InstructionType(int(math.Abs(rand.NormFloat64() + 1)) % 3)
		// Generate memory block from a normal distribution with std=8/3 and mean=8
		blockId := int(math.Abs(rand.NormFloat64() * 8/3 + 8)) % numMemoryBlocks

		// Create the instruction
		cpu.Instructions[i] = Instruction{InstructionType: iType, TargetBlockId: blockId}
		cpu.Instructions[i].Init()


		switch iType {

		case PROCESS:
			output.WriteString("PROCESS,0," + strconv.Itoa(blockId) + "\n")

		case READ:
			output.WriteString("READ,1," + strconv.Itoa(blockId) + "\n")

		case WRITE:
			output.WriteString("WRITE,2," + strconv.Itoa(blockId) + "\n")

		}
	}

	cpu.clock = clock
	cpu.Id = id
	file,_ := os.Create(instructionPath + cpuid)
	file.WriteString(output.String())
	file.Close()

	fmt.Println("iCPU",cpuid,"done")

}



func (cpu *CPU) executeInstruction(cache *CacheController, done *chan bool){

	instructions := cpu.Instructions[:]
	i := 0

	// for each instruction
	for len(instructions)  > 0{
		// Only do something if we are on high
		if <- *cpu.clock {

			// Select current instruction
			currentInstruction := instructions[0]
			// And remove it from the list
			instructions = instructions[1:]

			switch currentInstruction.InstructionType {

			case READ:
				cache.ReadBlock(currentInstruction.TargetBlockId)

			case WRITE:
				data := int(rand.NormFloat64() * 100)
				cache.WriteBlock(currentInstruction.TargetBlockId, data)
			default:
			}
			// Say we did one instruction to save state
			*done <- true

			// Mark the instruction as finished
			cpu.Instructions[i].IsFinished = true
			i++

			// Every instruction has a defined execution time
			time.Sleep(currentInstruction.ExecutionTime)
		}

	}

	close(*done)

}
