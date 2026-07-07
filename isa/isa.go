package isa

import "fmt"

//go:generate stringer -type=Register
type Register int32

//go:generate stringer -type=Opcode -trimprefix=Op
type Opcode int32

const (
	SP Register = iota
	RP
	X0
	X1
	X2
	X3
	X4
	T0
	T1
	T2
	T3
	T4
	A0
	A1
	A2
	ZERO
)

const (
	OpLW Opcode = iota
	OpSW
	OpLUI
	OpJZ
	OpMV
	OpINV
	OpADDI
	OpADD
	OpSUB
	OpMUL
	OpAND
	OpJ
	OpJR
	OpHALT
	OpDIV
	OpMOD
	OpJG
	OpJL
)

var BinaryToOpcode = map[int32]Opcode{
	0x0:  OpLW,
	0x1:  OpSW,
	0x2:  OpLUI,
	0x3:  OpJZ,
	0x4:  OpMV,
	0x5:  OpINV,
	0x6:  OpADDI,
	0x7:  OpADD,
	0x8:  OpSUB,
	0x9:  OpMUL,
	0xA:  OpAND,
	0xB:  OpJ,
	0xC:  OpJR,
	0xD:  OpHALT,
	0xE:  OpDIV,
	0xF:  OpMOD,
	0x10: OpJG,
	0x11: OpJL,
}

func (opcode Opcode) GetRead(r1, r2 int32) []int32 {
	var r []int32
	if _, ok := map[Opcode]struct{}{
		OpMV:   {},
		OpSW:   {},
		OpLW:   {},
		OpADDI: {},
		OpADD:  {},
		OpSUB:  {},
		OpMUL:  {},
		OpAND:  {},
		OpINV:  {},
		OpDIV:  {},
		OpMOD:  {},
		OpJR:   {},
		OpJZ:   {},
		OpJG:   {},
		OpJL:   {},
	}[opcode]; ok {
		r = append(r, r1)
	}
	if _, ok := map[Opcode]struct{}{
		OpSW:  {},
		OpADD: {},
		OpSUB: {},
		OpMUL: {},
		OpAND: {},
		OpDIV: {},
		OpMOD: {},
	}[opcode]; ok {
		r = append(r, r2)
	}
	return r
}

func (opcode Opcode) GetWrite(rd int32) int32 {
	if _, ok := map[Opcode]struct{}{
		OpLUI:  {},
		OpMV:   {},
		OpLW:   {},
		OpADDI: {},
		OpADD:  {},
		OpSUB:  {},
		OpMUL:  {},
		OpAND:  {},
		OpINV:  {},
		OpDIV:  {},
		OpMOD:  {},
	}[opcode]; ok {
		return rd
	}
	return -1
}

type ProgramItem struct {
	Address int
	Label   string
}

type Macros int

const (
	MacroNone Macros = iota
	MacroHi
	MacroLo
)

type Instruction struct {
	ProgramItem

	Opcode      Opcode
	Mnemonic    string
	Rs1         Register
	Rs2         Register
	Rd          Register
	Imm         int32
	Macros      Macros
	TargetLabel string
}

type Data struct {
	ProgramItem

	Value int32
}

type Program struct {
	Data         []Data
	Instructions []Instruction
}

func LUI(rd Register, imm int32) Instruction {
	return Instruction{
		Mnemonic: fmt.Sprintf("LUI %s, %d", rd, imm),
		Opcode:   OpLUI,
		Rd:       rd,
		Imm:      imm,
	}
}

func LUIHi(rd Register, label string) Instruction {
	return Instruction{
		Mnemonic:    fmt.Sprintf("LUI %s, %%hi(%s)", rd, label),
		Opcode:      OpLUI,
		Rd:          rd,
		TargetLabel: label,
		Macros:      MacroHi,
	}
}

func MV(rd, rs Register) Instruction {
	return Instruction{
		Mnemonic: fmt.Sprintf("MV %s, %s", rd, rs),
		Opcode:   OpMV,
		Rs1:      rs,
		Rd:       rd,
	}
}

func SW(rs2 Register, rs1 Register, imm int32) Instruction {
	return Instruction{
		Mnemonic: fmt.Sprintf("SW %s, %d(%s)", rs2, imm, rs1),
		Opcode:   OpSW,
		Rs1:      rs1,
		Rs2:      rs2,
		Imm:      imm,
	}
}

func LW(rd Register, rs Register, imm int32) Instruction {
	return Instruction{
		Mnemonic: fmt.Sprintf("LW %s, %d(%s)", rd, imm, rs),
		Opcode:   OpLW,
		Rs1:      rs,
		Rd:       rd,
		Imm:      imm,
	}
}

func ADDI(rd Register, rs Register, imm int32) Instruction {
	return Instruction{
		Mnemonic: fmt.Sprintf("ADDI %s, %s, %d", rd, rs, imm),
		Opcode:   OpADDI,
		Rs1:      rs,
		Rd:       rd,
		Imm:      imm,
	}
}

func ADDILo(rd Register, rs Register, label string) Instruction {
	return Instruction{
		Mnemonic:    fmt.Sprintf("ADDI %s, %s, %%lo(%s)", rd, rs, label),
		Opcode:      OpADDI,
		Rs1:         rs,
		Rd:          rd,
		TargetLabel: label,
		Macros:      MacroLo,
	}
}

func ADD(rd Register, rs1 Register, rs2 Register) Instruction {
	return Instruction{
		Mnemonic: fmt.Sprintf("ADD %s, %s, %s", rd, rs1, rs2),
		Opcode:   OpADD,
		Rs1:      rs1,
		Rs2:      rs2,
		Rd:       rd,
	}
}

func SUB(rd Register, rs1 Register, rs2 Register) Instruction {
	return Instruction{
		Mnemonic: fmt.Sprintf("SUB %s, %s, %s", rd, rs1, rs2),
		Opcode:   OpSUB,
		Rs1:      rs1,
		Rs2:      rs2,
		Rd:       rd,
	}
}

func MUL(rd Register, rs1 Register, rs2 Register) Instruction {
	return Instruction{
		Mnemonic: fmt.Sprintf("MUL %s, %s, %s", rd, rs1, rs2),
		Opcode:   OpMUL,
		Rs1:      rs1,
		Rs2:      rs2,
		Rd:       rd,
	}
}

func AND(rd Register, rs1 Register, rs2 Register) Instruction {
	return Instruction{
		Mnemonic: fmt.Sprintf("AND %s, %s, %s", rd, rs1, rs2),
		Opcode:   OpAND,
		Rs1:      rs1,
		Rs2:      rs2,
		Rd:       rd,
	}
}

func DIV(rd Register, rs1 Register, rs2 Register) Instruction {
	return Instruction{
		Mnemonic: fmt.Sprintf("DIV %s, %s, %s", rd, rs1, rs2),
		Opcode:   OpDIV,
		Rs1:      rs1,
		Rs2:      rs2,
		Rd:       rd,
	}
}

func MOD(rd Register, rs1 Register, rs2 Register) Instruction {
	return Instruction{
		Mnemonic: fmt.Sprintf("MOD %s, %s, %s", rd, rs1, rs2),
		Opcode:   OpMOD,
		Rs1:      rs1,
		Rs2:      rs2,
		Rd:       rd,
	}
}

func INV(rd Register, rs Register) Instruction {
	return Instruction{
		Mnemonic: fmt.Sprintf("INV %s, %s", rd, rs),
		Opcode:   OpINV,
		Rs1:      rs,
		Rd:       rd,
	}
}

func J(target string) Instruction {
	return Instruction{
		Mnemonic:    fmt.Sprintf("J %s", target),
		Opcode:      OpJ,
		TargetLabel: target,
	}
}

func JR(rs Register) Instruction {
	return Instruction{
		Mnemonic: fmt.Sprintf("JR %s", rs),
		Opcode:   OpJR,
		Rs1:      rs,
	}
}

func JZ(rs Register, target string) Instruction {
	return Instruction{
		Mnemonic:    fmt.Sprintf("JZ %s, %s", rs, target),
		Opcode:      OpJZ,
		Rs1:         rs,
		TargetLabel: target,
	}
}

func JG(rs Register, target string) Instruction {
	return Instruction{
		Mnemonic:    fmt.Sprintf("JG %s, %s", rs, target),
		Opcode:      OpJG,
		Rs1:         rs,
		TargetLabel: target,
	}
}

func JL(rs Register, target string) Instruction {
	return Instruction{
		Mnemonic:    fmt.Sprintf("JL %s, %s", rs, target),
		Opcode:      OpJL,
		Rs1:         rs,
		TargetLabel: target,
	}
}

func HALT() Instruction {
	return Instruction{
		Mnemonic: "HALT",
		Opcode:   OpHALT,
	}
}
