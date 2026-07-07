package isa

import (
	"encoding/binary"
	"fmt"
	"strings"
)

func ToBytes(program Program) []byte {
	var binaryCode []byte

	for _, instruction := range program.Instructions {
		instBin := uint32(instruction.Opcode << 27)
		instBin |= uint32(instruction.Rd << 23)
		instBin |= uint32(instruction.Rs1 << 19)
		instBin |= uint32(instruction.Rs2 << 15)

		var immMask uint32
		switch instruction.Opcode {
		case OpLUI, OpJ, OpJZ, OpJG, OpJL:
			immMask = 0xFFFFF
		default:
			immMask = 0xFFF
		}
		instBin |= uint32(instruction.Imm) & immMask
		binaryCode = binary.BigEndian.AppendUint32(binaryCode, instBin)
	}

	for _, data := range program.Data {
		binaryCode = binary.BigEndian.AppendUint32(binaryCode, uint32(data.Value))
	}

	return binaryCode
}

func ToHex(program Program) []string {
	var dump []string

	header := fmt.Sprintf("%-25s | %-10s | %-10s | %s", "<label>", "<address>", "<HEXCODE>", "<mnemonic>")
	dump = append(dump, header)

	for _, instruction := range program.Instructions {
		labels := strings.Split(instruction.Label, ":")
		var userLabels []string
		for _, label := range labels {
			if !strings.HasPrefix(label, "_") {
				userLabels = append(userLabels, label)
			}
		}
		userLabelsStr := strings.Join(userLabels, ":")
		address := instruction.Address
		opcode := fmt.Sprintf("0x%02X", uint32(instruction.Opcode))
		mnemonic := instruction.Mnemonic
		line := fmt.Sprintf("%-25s | %-10d | %-10s | %s", userLabelsStr, address, opcode, mnemonic)
		dump = append(dump, line)
	}

	for _, data := range program.Data {
		label := data.Label
		if strings.HasPrefix(label, "_") {
			label = ""
		}
		address := data.Address
		hexCode := fmt.Sprintf("0x%08X", data.Value)
		mnemonic := "DATA"
		line := fmt.Sprintf("%-25s | %-10d | %-10s | %s", label, address, hexCode, mnemonic)
		dump = append(dump, line)
	}

	return dump
}
