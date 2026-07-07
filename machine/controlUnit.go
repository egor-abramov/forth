package machine

import (
	"forth/isa"
	"slices"
)

type Signal int

const (
	writeMem Signal = iota
	readMem
	writeReg
	latchPC
	latchIR1
	latchIR2
	latchAR
	setALUFlags
	selALULeftPC
	selALULeftRS1
	selALURightRS2
	selALURightImm
	aluOpPassRight
	aluOpPassLeft
	aluOpPlus
	aluOpMinus
	aluOpMul
	aluOpDiv
	aluOpMod
	aluOpAnd
	aluOpInv
	selImmMode12
	selImmMode16
	selImmMode20
	selImmModeU
	selMemAddrPC
	selMemAddrAR
	selRegSrMem
	selRegSrALU
	selNextPCInc
	selNextPCALU
)

type JmpMode int

const (
	seqJmp JmpMode = iota
	seqInc
	seqMap
	seqJmpZ
	seqJmpL
	seqJmpG
)

type signalSet map[Signal]struct{}

type MicroCommand struct {
	signals signalSet
	jmpMode JmpMode
	jmpAddr int32
}

type controlUnit struct {
	dp            *dataPath
	mpc1          int32
	mpc2          int32
	scalarMode    bool
	dispatchTable map[isa.Opcode]int32
	mpMemory      []MicroCommand
}

func initDispatchTable() map[isa.Opcode]int32 {
	return map[isa.Opcode]int32{
		isa.OpLUI:  0x1,
		isa.OpMV:   0x2,
		isa.OpSW:   0x3,
		isa.OpLW:   0x5,
		isa.OpADDI: 0x7,
		isa.OpADD:  0x8,
		isa.OpSUB:  0x9,
		isa.OpMUL:  0xA,
		isa.OpAND:  0xB,
		isa.OpINV:  0xC,
		isa.OpDIV:  0xD,
		isa.OpMOD:  0xE,
		isa.OpJ:    0xF,
		isa.OpJR:   0x10,
		isa.OpJZ:   0x11,
		isa.OpJG:   0x14,
		isa.OpJL:   0x16,
		isa.OpHALT: 0x18,
	}
}

func initMPMemory() []MicroCommand {
	return []MicroCommand{
		// FETCH
		MicroCommand{
			signals: signalSet{
				selMemAddrPC: {},
				readMem:      {},
				latchIR1:     {},
				latchIR2:     {},
			},
			jmpMode: seqMap,
		},

		// LUI
		MicroCommand{
			signals: signalSet{
				selNextPCInc:   {},
				latchPC:        {},
				selImmModeU:    {},
				selALURightImm: {},
				aluOpPassRight: {},
				writeReg:       {},
				selRegSrALU:    {},
			},
		},

		// MV
		MicroCommand{
			signals: signalSet{
				selNextPCInc:  {},
				latchPC:       {},
				selALULeftRS1: {},
				aluOpPassLeft: {},
				setALUFlags:   {},
				selRegSrALU:   {},
				writeReg:      {},
			},
		},

		// SW
		MicroCommand{
			signals: signalSet{
				selNextPCInc:   {},
				latchPC:        {},
				selALULeftRS1:  {},
				selALURightImm: {},
				aluOpPlus:      {},
				latchAR:        {},
				selImmMode12:   {},
			},
			jmpMode: seqInc,
		},
		MicroCommand{
			signals: signalSet{
				selMemAddrAR: {},
				writeMem:     {},
			},
		},

		// LW
		MicroCommand{
			signals: signalSet{
				selNextPCInc:   {},
				latchPC:        {},
				selALULeftRS1:  {},
				selALURightImm: {},
				aluOpPlus:      {},
				latchAR:        {},
				selImmMode12:   {},
			},
			jmpMode: seqInc,
		},
		MicroCommand{
			signals: signalSet{
				selMemAddrAR: {},
				readMem:      {},
				writeReg:     {},
				selRegSrMem:  {},
			},
		},

		// ADDI
		MicroCommand{
			signals: signalSet{
				selNextPCInc:   {},
				latchPC:        {},
				selALULeftRS1:  {},
				selALURightImm: {},
				selImmMode12:   {},
				aluOpPlus:      {},
				setALUFlags:    {},
				writeReg:       {},
				selRegSrALU:    {},
			},
		},

		// ADD
		MicroCommand{
			signals: signalSet{
				selNextPCInc:   {},
				latchPC:        {},
				selALULeftRS1:  {},
				selALURightRS2: {},
				aluOpPlus:      {},
				setALUFlags:    {},
				writeReg:       {},
				selRegSrALU:    {},
			},
		},

		// SUB
		MicroCommand{
			signals: signalSet{
				selNextPCInc:   {},
				latchPC:        {},
				selALULeftRS1:  {},
				selALURightRS2: {},
				aluOpMinus:     {},
				setALUFlags:    {},
				writeReg:       {},
				selRegSrALU:    {},
			},
		},

		// MUL
		MicroCommand{
			signals: signalSet{
				selNextPCInc:   {},
				latchPC:        {},
				selALULeftRS1:  {},
				selALURightRS2: {},
				aluOpMul:       {},
				setALUFlags:    {},
				writeReg:       {},
				selRegSrALU:    {},
			},
		},

		// AND
		MicroCommand{
			signals: signalSet{
				selNextPCInc:   {},
				latchPC:        {},
				selALULeftRS1:  {},
				selALURightRS2: {},
				aluOpAnd:       {},
				setALUFlags:    {},
				writeReg:       {},
				selRegSrALU:    {},
			},
		},

		// INV
		MicroCommand{
			signals: signalSet{
				selNextPCInc:   {},
				latchPC:        {},
				selALULeftRS1:  {},
				selALURightRS2: {},
				aluOpInv:       {},
				setALUFlags:    {},
				writeReg:       {},
				selRegSrALU:    {},
			},
		},

		// DIV
		MicroCommand{
			signals: signalSet{
				selNextPCInc:   {},
				latchPC:        {},
				selALULeftRS1:  {},
				selALURightRS2: {},
				aluOpDiv:       {},
				setALUFlags:    {},
				writeReg:       {},
				selRegSrALU:    {},
			},
		},

		// MOD
		MicroCommand{
			signals: signalSet{
				selNextPCInc:   {},
				latchPC:        {},
				selALULeftRS1:  {},
				selALURightRS2: {},
				aluOpMod:       {},
				setALUFlags:    {},
				writeReg:       {},
				selRegSrALU:    {},
			},
		},

		// J
		MicroCommand{
			signals: signalSet{
				selALURightImm: {},
				aluOpPassRight: {},
				selNextPCALU:   {},
				latchPC:        {},
			},
		},

		// JR
		MicroCommand{
			signals: signalSet{
				selALULeftRS1: {},
				aluOpPassLeft: {},
				selNextPCALU:  {},
				latchPC:       {},
			},
		},

		// JZ
		MicroCommand{
			signals: signalSet{
				selALULeftRS1: {},
				aluOpPassLeft: {},
				setALUFlags:   {},
			},
			jmpMode: seqJmpZ,
			jmpAddr: 0x13,
		},
		MicroCommand{
			signals: signalSet{
				selNextPCInc: {},
				latchPC:      {},
			},
		},
		MicroCommand{
			signals: signalSet{
				selALURightImm: {},
				aluOpPassRight: {},
				selNextPCALU:   {},
				latchPC:        {},
			},
		},

		// JG
		MicroCommand{
			signals: signalSet{
				selALULeftRS1: {},
				aluOpPassLeft: {},
				setALUFlags:   {},
			},
			jmpMode: seqJmpG,
			jmpAddr: 0x13,
		},
		MicroCommand{
			signals: signalSet{
				selNextPCInc: {},
				latchPC:      {},
			},
		},

		// JL
		MicroCommand{
			signals: signalSet{
				selALULeftRS1: {},
				aluOpPassLeft: {},
				setALUFlags:   {},
			},
			jmpMode: seqJmpL,
			jmpAddr: 0x13,
		},
		MicroCommand{
			signals: signalSet{
				selNextPCInc: {},
				latchPC:      {},
			},
		},

		// HALT
		MicroCommand{
			jmpAddr: 0x18,
		},

		// NOP
		MicroCommand{},
	}
}

func newControlUnit(dp *dataPath) *controlUnit {
	return &controlUnit{
		dp:            dp,
		dispatchTable: initDispatchTable(),
		mpMemory:      initMPMemory(),
	}
}

func (cu *controlUnit) hazardsResolve(ir1, ir2 int32) bool {
	if cu.scalarMode {
		return false
	}

	opcode1 := isa.BinaryToOpcode[(ir1>>27)&0x1F]
	opcode2, ok2 := isa.BinaryToOpcode[(ir2>>27)&0x1F]

	branches := []isa.Opcode{isa.OpJ, isa.OpJR, isa.OpJZ, isa.OpJG, isa.OpJL, isa.OpHALT}
	memOps := []isa.Opcode{isa.OpLW, isa.OpSW}

	if !ok2 {
		return false
	}

	if slices.Contains(branches, opcode1) || slices.Contains(branches, opcode2) {
		return false
	}

	if opcode1 == isa.OpHALT || opcode2 == isa.OpHALT {
		return false
	}

	if slices.Contains(memOps, opcode1) && slices.Contains(memOps, opcode2) {
		return false
	}

	rd1 := (ir1 >> 23) & 0xF
	rs21 := (ir1 >> 15) & 0xF

	rd2 := (ir2 >> 23) & 0xF
	rs12 := (ir2 >> 19) & 0xF
	rs22 := (ir2 >> 15) & 0xF

	w1 := opcode1.GetWrite(rd1)
	w2 := opcode2.GetWrite(rd2)
	r2 := opcode2.GetRead(rs12, rs22)

	if w1 != -1 && w1 != cu.dp.regFile.zeroRegisterAddr {
		if slices.Contains(r2, w1) || w1 == w2 {
			return false
		}
	}

	if w2 != -1 && w2 != cu.dp.regFile.zeroRegisterAddr {
		if opcode1 == isa.OpSW && w2 == rs21 {
			return false
		}
	}

	return true
}

func (cu *controlUnit) nextMPC(mpc int32, mc MicroCommand) int32 {
	jmpAddr := mc.jmpAddr
	opcode := isa.BinaryToOpcode[(cu.dp.ir1>>27)&0x1F]

	switch mc.jmpMode {
	case seqInc:
		return mpc + 1
	case seqMap:
		addr := cu.dispatchTable[opcode]
		return addr
	case seqJmp:
		return jmpAddr
	case seqJmpZ:
		if f := cu.dp.flags["Z"]; f == 1 {
			return jmpAddr
		}
		return mpc + 1
	case seqJmpG:
		n := cu.dp.flags["N"]
		z := cu.dp.flags["Z"]
		if n != 1 && z != 1 {
			return jmpAddr
		}
		return mpc + 1
	case seqJmpL:
		n := cu.dp.flags["N"]
		z := cu.dp.flags["Z"]
		if n == 1 && z != 1 {
			return jmpAddr
		}
		return mpc + 1
	}
	return 0
}

func (cu *controlUnit) tick() (bool, error) {
	if cu.mpc1 == cu.dispatchTable[isa.OpHALT] {
		return false, nil
	}

	// sync logic
	var mc1, mc2 MicroCommand
	if cu.mpc1 == 0 && cu.mpc2 != 0 {
		mc1 = cu.mpMemory[0x19]
		mc2 = cu.mpMemory[cu.mpc2]
	} else if cu.mpc1 != 0 && cu.mpc2 == 0 {
		mc2 = cu.mpMemory[0x19]
		mc1 = cu.mpMemory[cu.mpc1]
	} else {
		mc1 = cu.mpMemory[cu.mpc1]
		mc2 = cu.mpMemory[cu.mpc2]
	}

	err := cu.dp.tick(mc1.signals, mc2.signals)
	if err != nil {
		return false, err
	}
	if mc1.jmpMode == seqMap && cu.mpc1 == 0 && cu.mpc2 == 0 {
		op1 := isa.BinaryToOpcode[(cu.dp.ir1>>27)&0x1F]
		op2 := isa.BinaryToOpcode[(cu.dp.ir2>>27)&0x1F]

		op1Addr, ok1 := cu.dispatchTable[op1]
		if !ok1 {
			op1Addr = 0x18
		}
		cu.mpc1 = op1Addr
		if cu.hazardsResolve(cu.dp.ir1, cu.dp.ir2) {
			op2Addr, ok2 := cu.dispatchTable[op2]
			if !ok2 {
				op2Addr = 0x18
			}
			cu.mpc2 = op2Addr
		} else {
			cu.mpc2 = 0x0
		}
	} else {
		cu.mpc1 = cu.nextMPC(cu.mpc1, mc1)
		cu.mpc2 = cu.nextMPC(cu.mpc2, mc2)
	}

	return true, nil
}
