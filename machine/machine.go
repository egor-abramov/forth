package machine

import (
	"fmt"
	"forth/isa"
	"log"
	"os"
	"strings"
)

func initMachine(sourcePath, inputPath string) ([]int32, []int32, error) {
	initialMemBytes, err := os.ReadFile(sourcePath)
	if err != nil {
		log.Fatal(err)
	}

	initialMem := make([]int32, len(initialMemBytes))
	for i, b := range initialMemBytes {
		initialMem[i] = int32(b)
	}

	dataBytes, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, nil, err
	}
	data := string(dataBytes)
	inputTokens := make([]int32, 0, len(data)+1)

	for _, ch := range data {
		inputTokens = append(inputTokens, ch)
	}

	inputTokens = append(inputTokens, '\n')
	return initialMem, inputTokens, nil
}

func Simulate(sourcePath, inputPath string, trace bool, scalarMode bool) {
	initialMemory, inputTokens, err := initMachine(sourcePath, inputPath)
	if err != nil {
		log.Fatal(err)
	}

	mem := newMemory(initialMemory, inputTokens)
	dp := newDataPath(mem)
	cu := newControlUnit(dp)

	var ticks int
	var traceLog []string

	if scalarMode {
		log.Println("superscalar mode disabled")
	}

	for {
		ticks++
		isRunning, err := cu.tick()

		if err != nil {
			for _, traceLine := range traceLog {
				log.Println(traceLine)
			}
			log.Fatal(err)
		}
		if !isRunning {
			break
		}

		if trace && len(traceLog) < 1000 {
			pcStr := fmt.Sprintf("PC: 0x%04X", dp.pc)
			mpcStr := fmt.Sprintf("m1: %d m2: %d", cu.mpc1, cu.mpc2)

			op1, ok1 := isa.BinaryToOpcode[(dp.ir1>>27)&0x1F]
			op2, ok2 := isa.BinaryToOpcode[(dp.ir2>>27)&0x1F]

			var opcode1Str, opcode2Str string
			if ok1 && cu.mpc1 != 0x0 && cu.mpc1 != 0x19 {
				opcode1Str = op1.String()
			} else {
				opcode1Str = "IDLE"
			}

			if ok2 && cu.mpc2 != 0x0 && cu.mpc2 != 0x19 {
				opcode2Str = op2.String()
			} else {
				opcode2Str = "IDLE"
			}

			if cu.mpc1 == 0x0 && cu.mpc2 == 0x0 {
				opcode1Str = "FETCH"
				opcode2Str = "FETCH"
			}

			flagsStr := fmt.Sprintf("Z:%d N:%d", dp.flags["Z"], dp.flags["N"])

			traceLine := fmt.Sprintf("Tick: %04d | %-12s | %-12s | %5s | %5s | %s", ticks, pcStr, mpcStr, opcode1Str, opcode2Str, flagsStr)
			traceLog = append(traceLog, traceLine)
		}
	}

	log.Printf("Ticks executed: %d\n", ticks)
	log.Println("Output:")
	output := strings.Join(mem.outputBuffer, "") + "\n"
	log.Println(output)
	if trace {
		log.Println("Trace:")
		for _, traceLine := range traceLog {
			log.Println(traceLine)
		}
	}
}
