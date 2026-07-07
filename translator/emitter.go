package translator

import (
	"fmt"
	"forth/isa"
)

type Emitter struct {
	curAddr  int
	curLabel string

	program isa.Program
}

func NewEmitter() *Emitter {
	return &Emitter{
		curLabel: "",
		program:  isa.Program{},
	}
}

func (e *Emitter) emit(instruction isa.Instruction) {
	if e.curLabel != "" {
		instruction.Label = e.curLabel
		e.curLabel = ""
	}
	instruction.Address = e.curAddr
	e.program.Instructions = append(e.program.Instructions, instruction)
	e.curAddr += wordSize
}

func (e *Emitter) emitData(data isa.Data) {
	e.program.Data = append(e.program.Data, data)
}

func (e *Emitter) emitLabel(label string) {
	if e.curLabel != "" {
		e.curLabel += ":"
	}
	e.curLabel += label
}

func (e *Emitter) emitLit(val any) {
	e.emit(isa.ADDI(isa.SP, isa.SP, -4))
	e.emit(isa.SW(isa.T1, isa.SP, 0))
	e.emit(isa.MV(isa.T1, isa.T0))

	switch v := val.(type) {
	case int:
		e.emitLoadImmNum(isa.T0, int32(v))
	case int32:
		e.emitLoadImmNum(isa.T0, v)
	case string:
		e.emitLoadImmLabel(isa.T0, v)
	}
}

func (e *Emitter) emitLoadImmNum(reg isa.Register, val int32) {
	if val >= -2048 && val <= 2047 {
		e.emit(isa.ADDI(reg, isa.ZERO, val))
	} else {
		upper := (val >> 12) & 0xFFFFF
		lower := val & 0xFFF
		if lower&0x800 != 0 {
			upper++
		}
		e.emit(isa.LUI(reg, upper))
		e.emit(isa.ADDI(reg, reg, lower))
	}
}

func (e *Emitter) emitLoadImmLabel(reg isa.Register, label string) {
	e.emit(isa.LUIHi(reg, label))
	e.emit(isa.ADDILo(reg, reg, label))
}

func (e *Emitter) emitCall(targetLabel string) {
	retLabel := fmt.Sprintf("_RET_%d", e.curAddr+5*wordSize)
	e.emitLoadImmLabel(isa.T2, retLabel)
	e.emit(isa.ADDI(isa.RP, isa.RP, -4))
	e.emit(isa.SW(isa.T2, isa.RP, 0))
	e.emit(isa.J(targetLabel))
	e.emitLabel(retLabel)
}

func (e *Emitter) emitRet() {
	e.emit(isa.LW(isa.T2, isa.RP, 0))
	e.emit(isa.ADDI(isa.RP, isa.RP, 4))
	e.emit(isa.JR(isa.T2))
}
