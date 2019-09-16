package Core

import "time"

// Memory delay
const writeDelay 	= time.Second * 3
const readDelay  	= time.Second * 3
const processDelay  = time.Second * 1

// Define enum to represent each State of the CacheController Block
type InstructionType int
const (  // iota is reset to 0
	PROCESS 	InstructionType = iota  // PROCESS = 0
	READ 		InstructionType = iota  // READ = 1
	WRITE		InstructionType = iota  // WRITE = 2
)

type Instruction struct{
	InstructionType InstructionType
	TargetBlockId   int
	ExecutionTime   time.Duration
	IsFinished		bool
}

func (instruction *Instruction) Init(){

	// Deifne its execution time
	switch instruction.InstructionType {

	case READ:
		instruction.ExecutionTime = readDelay

	case WRITE:
		instruction.ExecutionTime = writeDelay

	case PROCESS:
		instruction.ExecutionTime = processDelay

	}
}
