package machine

import (
	"fmt"
)

type registerFile struct {
	registerCount    int
	registerLen      int
	zeroRegisterAddr int32
	regs             []int32
}

func newRegisterFile() *registerFile {
	return &registerFile{
		registerCount:    16,
		registerLen:      32,
		zeroRegisterAddr: 0xF,
		regs:             make([]int32, 16),
	}
}

func (rf *registerFile) read(addr int32) int32 {
	if addr == rf.zeroRegisterAddr {
		return 0
	}
	return rf.regs[addr]
}

func (rf *registerFile) write(val int32, addr int32) {
	if addr == rf.zeroRegisterAddr {
		return
	}
	rf.regs[addr] = val
}

type memory struct {
	mem            []int32
	memSize        int32
	inputBuffer    []int32
	inputAddr      int32
	outputNumAddr  int32
	outputCharAddr int32
	outputBuffer   []string
}

func newMemory(initialMem []int32, inputBuffer []int32) *memory {
	memory := memory{
		memSize:        0x4000,
		inputAddr:      0x3EF8,
		outputNumAddr:  0x3EFC,
		outputCharAddr: 0x3F00,
		inputBuffer:    inputBuffer,
	}

	mem := make([]int32, memory.memSize)
	copy(mem, initialMem)
	memory.mem = mem

	return &memory
}

func (m *memory) read(addr int32) (int32, int32, error) {
	switch {
	case addr == m.inputAddr:
		if len(m.inputBuffer) == 0 {
			return 0, 0, fmt.Errorf("no elements in input buffer")
		}
		el := m.inputBuffer[0]
		m.inputBuffer = m.inputBuffer[1:]
		return el, 0, nil

	case addr >= 0 &&
		addr < m.memSize-3 && addr != m.outputNumAddr &&
		addr != m.outputCharAddr:

		var (
			firstWord  int32
			secondWord int32
		)
		if addr+4 < m.memSize-3 {
			secondWord = (m.mem[addr+4] << 24) |
				(m.mem[addr+5] << 16) |
				(m.mem[addr+6] << 8) |
				m.mem[addr+7]
		}

		firstWord = (m.mem[addr] << 24) |
			(m.mem[addr+1] << 16) |
			(m.mem[addr+2] << 8) |
			m.mem[addr+3]

		return firstWord, secondWord, nil
	default:
		return 0, 0, fmt.Errorf("invalid memory access at address %d", addr)
	}
}

func (m *memory) write(val int32, addr int32) error {
	switch {
	case addr == m.outputNumAddr:
		m.outputBuffer = append(m.outputBuffer, fmt.Sprintf("%d ", val))
	case addr == m.outputCharAddr:
		m.outputBuffer = append(m.outputBuffer, fmt.Sprintf("%c", val))
	case addr >= 0 && addr < m.memSize-3 && addr != m.inputAddr:
		m.mem[addr] = (val >> 24) & 0xFF
		m.mem[addr+1] = (val >> 16) & 0xFF
		m.mem[addr+2] = (val >> 8) & 0xFF
		m.mem[addr+3] = val & 0xFF
	default:
		return fmt.Errorf("invalid memory access at address %d", addr)
	}
	return nil
}

type dataPath struct {
	memory  *memory
	regFile *registerFile
	pc      int32
	ir1     int32
	ir2     int32
	ar1     int32
	ar2     int32
	flags   map[string]int32
}

func newDataPath(memory *memory) *dataPath {
	return &dataPath{
		memory:  memory,
		regFile: newRegisterFile(),
		flags:   make(map[string]int32),
	}
}

func (dp *dataPath) tick(signals1, signals2 signalSet) error {
	var err error

	// decode ir1
	rd1Idx := (dp.ir1 >> 23) & 0xF
	rs11Idx := (dp.ir1 >> 19) & 0xF
	rs21Idx := (dp.ir1 >> 15) & 0xF
	imm1Ir := dp.ir1 & 0xFFFFF

	// decode ir2
	rd2Idx := (dp.ir2 >> 23) & 0xF
	rs12Idx := (dp.ir2 >> 19) & 0xF
	rs22Idx := (dp.ir2 >> 15) & 0xF
	imm2Ir := dp.ir2 & 0xFFFFF

	rs11Data := dp.regFile.read(rs11Idx)
	rs21Data := dp.regFile.read(rs21Idx)
	rs12Data := dp.regFile.read(rs12Idx)
	rs22Data := dp.regFile.read(rs22Idx)

	imm1Data := dp.immGenerator(imm1Ir, signals1)
	imm2Data := dp.immGenerator(imm2Ir, signals2)

	// ALU 1
	var alu1L, alu1R int32
	if _, ok := signals1[selALULeftPC]; ok {
		alu1L = dp.pc
	}
	if _, ok := signals1[selALULeftRS1]; ok {
		alu1L = rs11Data
	}

	if _, ok := signals1[selALURightImm]; ok {
		alu1R = imm1Data
	}
	if _, ok := signals1[selALURightRS2]; ok {
		alu1R = rs21Data
	}
	alu1Res, flags1 := dp.aluExecute(alu1L, alu1R, signals1)

	// ALU 2
	var alu2L, alu2R int32
	if _, ok := signals2[selALULeftPC]; ok {
		alu2L = dp.pc
	}
	if _, ok := signals2[selALULeftRS1]; ok {
		alu2L = rs12Data
	}

	if _, ok := signals2[selALURightImm]; ok {
		alu2R = imm2Data
	}
	if _, ok := signals2[selALURightRS2]; ok {
		alu2R = rs22Data
	}
	alu2Res, flags2 := dp.aluExecute(alu2L, alu2R, signals2)

	// set flags
	if _, ok := signals2[setALUFlags]; ok {
		dp.flags = flags2
	} else if _, ok := signals1[setALUFlags]; ok {
		dp.flags = flags1
	}

	// select memory address source
	var memAddr int32
	if _, ok := signals1[selMemAddrAR]; ok {
		memAddr = dp.ar1
	} else if _, ok := signals2[selMemAddrAR]; ok {
		memAddr = dp.ar2
	} else if _, ok := signals1[selMemAddrPC]; ok {
		memAddr = dp.pc
	} else if _, ok := signals2[selMemAddrPC]; ok {
		memAddr = dp.pc
	}

	// write to memory
	if _, ok := signals1[writeMem]; ok {
		err = dp.memory.write(rs21Data, memAddr)
		if err != nil {
			return err
		}
	} else if _, ok := signals2[writeMem]; ok {
		err = dp.memory.write(rs22Data, memAddr)
		if err != nil {
			return err
		}
	}

	// read from memory
	var memDataOut1, memDataOut2 int32
	if _, ok := signals1[readMem]; ok {
		memDataOut1, memDataOut2, err = dp.memory.read(memAddr)
		if err != nil {
			return err
		}
	} else if _, ok := signals2[readMem]; ok {
		memDataOut1, memDataOut2, err = dp.memory.read(memAddr)
		if err != nil {
			return err
		}
	}

	// data to register
	var regWriteData1 int32
	if _, ok := signals1[selRegSrMem]; ok {
		regWriteData1 = memDataOut1
	} else if _, ok := signals1[selRegSrALU]; ok {
		regWriteData1 = alu1Res
	}

	var regWriteData2 int32
	if _, ok := signals2[selRegSrMem]; ok {
		regWriteData2 = memDataOut1
	} else if _, ok := signals2[selRegSrALU]; ok {
		regWriteData2 = alu2Res
	}

	// write to register
	if _, ok := signals1[writeReg]; ok {
		dp.regFile.write(regWriteData1, rd1Idx)
	}
	if _, ok := signals2[writeReg]; ok {
		dp.regFile.write(regWriteData2, rd2Idx)
	}

	// latch AR
	if _, ok := signals1[latchAR]; ok {
		dp.ar1 = alu1Res
	}
	if _, ok := signals2[latchAR]; ok {
		dp.ar2 = alu2Res
	}

	// latch IR
	if _, ok := signals1[latchIR1]; ok {
		dp.ir1 = memDataOut1
	}
	if _, ok := signals2[latchIR1]; ok {
		dp.ir1 = memDataOut1
	}
	if _, ok := signals1[latchIR2]; ok {
		dp.ir2 = memDataOut2
	}
	if _, ok := signals2[latchIR2]; ok {
		dp.ir2 = memDataOut2
	}

	// increment PC
	var pcInc, newPC int32
	if _, ok := signals1[selNextPCInc]; ok {
		pcInc += 4
	}
	if _, ok := signals2[selNextPCInc]; ok {
		pcInc += 4
	}

	if _, ok := signals1[selNextPCALU]; ok {
		newPC = alu1Res
	} else if _, ok := signals2[selNextPCALU]; ok {
		newPC = alu2Res
	} else {
		newPC = dp.pc + pcInc
	}

	// latch pc
	if _, ok := signals1[latchPC]; ok {
		dp.pc = newPC
	}
	if _, ok := signals2[latchPC]; ok {
		dp.pc = newPC
	}

	return nil
}

func (dp *dataPath) immGenerator(val int32, signals signalSet) int32 {
	var n int32
	if _, ok := signals[selImmMode12]; ok {
		val &= 0xFFF
		n = 12
	} else if _, ok := signals[selImmMode16]; ok {
		val &= 0xFFFF
		n = 16
	} else if _, ok := signals[selImmMode20]; ok {
		val &= 0xFFFFF
		n = 20
	} else if _, ok := signals[selImmModeU]; ok {
		return (val & 0xFFFFF) << 12
	} else {
		return val & 0x7FFFF
	}
	signBit := int32(1) << (n - 1)
	return (val & (signBit - 1)) - (val & signBit)
}

func (dp *dataPath) aluExecute(aluL int32, aluR int32, signals signalSet) (int32, map[string]int32) {
	var res int32
	if _, ok := signals[aluOpPlus]; ok {
		res = aluL + aluR
	} else if _, ok := signals[aluOpMinus]; ok {
		res = aluL - aluR
	} else if _, ok := signals[aluOpMul]; ok {
		res = aluL * aluR
	} else if _, ok := signals[aluOpAnd]; ok {
		res = aluL & aluR
	} else if _, ok := signals[aluOpInv]; ok {
		res = ^aluL
	} else if _, ok := signals[aluOpPassLeft]; ok {
		res = aluL
	} else if _, ok := signals[aluOpPassRight]; ok {
		res = aluR
	} else if _, ok := signals[aluOpDiv]; ok {
		if aluR == 0 {
			res = 0
		} else {
			res = aluL / aluR
			if (aluL < 0) != (aluR < 0) && aluL%aluR != 0 {
				res--
			}
		}
	} else if _, ok := signals[aluOpMod]; ok {
		if aluR == 0 {
			res = 0
		} else {
			res = aluL % aluR
			if (aluL < 0) != (aluR < 0) && res != 0 {
				res += aluR
			}
		}
	}
	flags := make(map[string]int32)
	if (res>>31)&1 == 1 {
		flags["N"] = 1
	}
	if res == 0 {
		flags["Z"] = 1
	}

	return res, flags
}
