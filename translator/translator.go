package translator

import (
	"fmt"
	"forth/isa"
	"log"
	"os"
	"strings"
	"unicode/utf8"
)

const (
	wordSize            = 4
	dataStackInitAddr   = 0x3F80
	returnStackInitAddr = 0x4000
	inputAddr           = 0x3EF8
	outputNumAddr       = 0x3EFC
	outputCharAddr      = 0x3F00
)

type Contract = struct {
	Takes    int32
	Puts     int32
	IsUnsafe bool
}

type IfStackElem = struct {
	tag         string
	label       string
	depthBefore int32
	depthAfter  int32
}

var (
	var2Label = map[string]string{}
	word2addr = map[string]int{}

	loopStack  []string
	ifStack    []IfStackElem
	funcSkips  []string
	lastNumber int32 = 0

	funcContracts = map[string]Contract{}

	currentDepth    int32
	savedDepth      int32
	expectedPuts    int32
	isCurrentUnsafe bool
	currentProc     string
)

func translateTokens(tokens []token, emitter *Emitter) error {
	i := 0
	for i < len(tokens) {
		token := tokens[i]
		isSuccess, nextI := peephole(tokens, i, emitter)
		i = nextI
		if isSuccess {
			continue
		}
		switch token.typ {
		case "WORD":
			nextI, err := translateWord(i, tokens, emitter)
			if err != nil {
				return err
			}
			i = nextI
		case "NUMBER":
			lastNumber = token.valNum
			emitter.emitLit(token.valNum)
			currentDepth++
		}
		i++
	}

	return nil
}

func translateWord(index int, tokens []token, emitter *Emitter) (int, error) {
	word := tokens[index].valStr

	switch word {
	case "+":
		emitter.emit(isa.ADD(isa.T0, isa.T1, isa.T0))
		emitter.emit(isa.LW(isa.T1, isa.SP, 0))
		emitter.emit(isa.ADDI(isa.SP, isa.SP, 4))
		currentDepth--
	case "-":
		emitter.emit(isa.SUB(isa.T0, isa.T1, isa.T0))
		emitter.emit(isa.LW(isa.T1, isa.SP, 0))
		emitter.emit(isa.ADDI(isa.SP, isa.SP, 4))
		currentDepth--
	case "*":
		emitter.emit(isa.MUL(isa.T0, isa.T1, isa.T0))
		emitter.emit(isa.LW(isa.T1, isa.SP, 0))
		emitter.emit(isa.ADDI(isa.SP, isa.SP, 4))
		currentDepth--
	case "/":
		emitter.emit(isa.DIV(isa.T0, isa.T1, isa.T0))
		emitter.emit(isa.LW(isa.T1, isa.SP, 0))
		emitter.emit(isa.ADDI(isa.SP, isa.SP, 4))
	case "%":
		emitter.emit(isa.MOD(isa.T0, isa.T1, isa.T0))
		emitter.emit(isa.LW(isa.T1, isa.SP, 0))
		emitter.emit(isa.ADDI(isa.SP, isa.SP, 4))
		currentDepth--
	case "AND":
		emitter.emit(isa.AND(isa.T0, isa.T1, isa.T0))
		emitter.emit(isa.LW(isa.T1, isa.SP, 0))
		emitter.emit(isa.ADDI(isa.SP, isa.SP, 4))
		currentDepth--
	case "NOT":
		emitter.emit(isa.INV(isa.T0, isa.T0))
	case "DUP":
		emitter.emit(isa.ADDI(isa.SP, isa.SP, -4))
		emitter.emit(isa.SW(isa.T1, isa.SP, 0))
		emitter.emit(isa.MV(isa.T1, isa.T0))
		currentDepth++
	case "DROP":
		emitter.emit(isa.MV(isa.T0, isa.T1))
		emitter.emit(isa.LW(isa.T1, isa.SP, 0))
		emitter.emit(isa.ADDI(isa.SP, isa.SP, 4))
		currentDepth--
	case "SWAP":
		emitter.emit(isa.MV(isa.T2, isa.T0))
		emitter.emit(isa.MV(isa.T0, isa.T1))
		emitter.emit(isa.MV(isa.T1, isa.T2))
	case "!":
		emitter.emit(isa.SW(isa.T1, isa.T0, 0))
		emitter.emit(isa.LW(isa.T0, isa.SP, 0))
		emitter.emit(isa.LW(isa.T1, isa.SP, 4))
		emitter.emit(isa.ADDI(isa.SP, isa.SP, 8))
		currentDepth -= 2
	case "@":
		emitter.emit(isa.LW(isa.T0, isa.T0, 0))
	case "READ":
		emitter.emit(isa.ADDI(isa.SP, isa.SP, -4))
		emitter.emit(isa.SW(isa.T1, isa.SP, 0))
		emitter.emit(isa.MV(isa.T1, isa.T0))
		emitter.emitLoadImmNum(isa.T2, inputAddr)
		emitter.emit(isa.LW(isa.T0, isa.T2, 0))
		currentDepth++
	case ".":
		emitter.emitLoadImmNum(isa.T2, outputNumAddr)
		emitter.emit(isa.SW(isa.T0, isa.T2, 0))
		emitter.emit(isa.MV(isa.T0, isa.T1))
		emitter.emit(isa.LW(isa.T1, isa.SP, 0))
		emitter.emit(isa.ADDI(isa.SP, isa.SP, 4))
		currentDepth--
	case "EMIT":
		emitter.emitLoadImmNum(isa.T2, outputCharAddr)
		emitter.emit(isa.SW(isa.T0, isa.T2, 0))
		emitter.emit(isa.MV(isa.T0, isa.T1))
		emitter.emit(isa.LW(isa.T1, isa.SP, 0))
		emitter.emit(isa.ADDI(isa.SP, isa.SP, 4))
		currentDepth--
	case "LOOP":
		loopLabel := fmt.Sprintf("LOOP_%d", emitter.curAddr)
		loopStack = append(loopStack, loopLabel)
		emitter.emitLabel(loopLabel)
	case "ENDLOOP":
		if len(loopStack) == 0 {
			return index, fmt.Errorf("syntax error: loop expected")
		}
		last := len(loopStack) - 1
		beginLoopLabel := loopStack[last]
		loopStack = loopStack[:last]
		endLoopLabel := fmt.Sprintf("ENDLOOP_%d", emitter.curAddr)
		emitter.emit(isa.MV(isa.T2, isa.T0))
		emitter.emit(isa.MV(isa.T0, isa.T1))
		emitter.emit(isa.LW(isa.T1, isa.SP, 0))
		emitter.emit(isa.ADDI(isa.SP, isa.SP, 4))
		emitter.emit(isa.JZ(isa.T2, endLoopLabel))
		emitter.emit(isa.J(beginLoopLabel))
		emitter.emitLabel(endLoopLabel)
		currentDepth--
	case "IF":
		falseLabel := fmt.Sprintf("IF_FALSE_%d", emitter.curAddr)
		currentDepth--
		ifElem := IfStackElem{
			tag:         "if",
			label:       falseLabel,
			depthBefore: currentDepth,
		}
		ifStack = append(ifStack, ifElem)

		emitter.emit(isa.MV(isa.T2, isa.T0))
		emitter.emit(isa.MV(isa.T0, isa.T1))
		emitter.emit(isa.LW(isa.T1, isa.SP, 0))
		emitter.emit(isa.ADDI(isa.SP, isa.SP, 4))
		emitter.emit(isa.JZ(isa.T2, falseLabel))
	case "ELSE":
		if len(ifStack) == 0 {
			return index, fmt.Errorf("syntax error: if expected")
		}
		last := len(ifStack) - 1
		ifElem := ifStack[last]
		ifStack = ifStack[:last]

		if ifElem.tag != "if" {
			return index, fmt.Errorf("syntax error: if expected")
		}
		ifElem.depthAfter = currentDepth
		currentDepth = ifElem.depthBefore

		endLabel := fmt.Sprintf("IF_END_%d", emitter.curAddr)
		emitter.emit(isa.J(endLabel))
		emitter.emitLabel(ifElem.label)
		ifElem.tag = "else"
		ifElem.label = endLabel
		ifStack = append(ifStack, ifElem)
	case "THEN":
		last := len(ifStack) - 1
		ifElem := ifStack[last]
		ifStack = ifStack[:last]

		if ifElem.tag == "else" {
			if currentDepth != ifElem.depthAfter && !isCurrentUnsafe {
				return index, fmt.Errorf("branch stack mismatch: IF ended with %d, ELSE ended with %d", ifElem.depthAfter, currentDepth)
			}
		} else if ifElem.tag == "if" {
			if currentDepth != ifElem.depthBefore && !isCurrentUnsafe {
				return index, fmt.Errorf("branch stack mismatch: IF block changed stack depth without ELSE")
			}
		} else {
			return index, fmt.Errorf("syntax error: if expected")
		}
		emitter.emitLabel(ifElem.label)
	case "=0":
		trueLabel := fmt.Sprintf("IS_TRUE_%d", emitter.curAddr)
		falseLabel := fmt.Sprintf("IS_FALSE_%d", emitter.curAddr)
		emitter.emit(isa.MV(isa.T2, isa.T0))
		emitter.emit(isa.ADDI(isa.T0, isa.ZERO, 0))
		emitter.emit(isa.JZ(isa.T2, trueLabel))
		emitter.emit(isa.J(falseLabel))
		emitter.emitLabel(trueLabel)
		emitter.emit(isa.ADDI(isa.T0, isa.ZERO, 1))
		emitter.emitLabel(falseLabel)
	case ">0":
		trueLabel := fmt.Sprintf("IS_TRUE_%d", emitter.curAddr)
		falseLabel := fmt.Sprintf("IS_FALSE_%d", emitter.curAddr)
		emitter.emit(isa.MV(isa.T2, isa.T0))
		emitter.emit(isa.ADDI(isa.T0, isa.ZERO, 0))
		emitter.emit(isa.JG(isa.T2, trueLabel))
		emitter.emit(isa.J(falseLabel))
		emitter.emitLabel(trueLabel)
		emitter.emit(isa.ADDI(isa.T0, isa.ZERO, 1))
		emitter.emitLabel(falseLabel)
	case "<0":
		trueLabel := fmt.Sprintf("IS_TRUE_%d", emitter.curAddr)
		falseLabel := fmt.Sprintf("IS_FALSE_%d", emitter.curAddr)
		emitter.emit(isa.MV(isa.T2, isa.T0))
		emitter.emit(isa.ADDI(isa.T0, isa.ZERO, 0))
		emitter.emit(isa.JL(isa.T2, trueLabel))
		emitter.emit(isa.J(falseLabel))
		emitter.emitLabel(trueLabel)
		emitter.emit(isa.ADDI(isa.T0, isa.ZERO, 1))
		emitter.emitLabel(falseLabel)
	case ":":
		index++
		nextToken := tokens[index]
		tokenName := strings.ToUpper(nextToken.valStr)

		if index+1 < len(tokens) && tokens[index+1].typ == "CONTRACT" {
			index++
			contractToken := tokens[index]

			isUnsafe := false
			for j := index; j < len(tokens); j++ {
				if tokens[j].typ == "WORD" && tokens[j].valStr == "EXECUTE" {
					isUnsafe = true
				}
				if tokens[j].valStr == ";" {
					break
				}
			}

			funcContracts[tokenName] = Contract{
				Puts:     contractToken.puts,
				Takes:    contractToken.takes,
				IsUnsafe: isUnsafe,
			}

			savedDepth = currentDepth
			currentDepth = contractToken.takes
			expectedPuts = contractToken.puts
			isCurrentUnsafe = isUnsafe
			currentProc = tokenName
		} else {
			return index, fmt.Errorf("syntax error: contract expected in %s procedure", tokenName)
		}

		skipLabel := fmt.Sprintf("_SKIP_FUNC_%s", tokenName)
		emitter.emit(isa.J(skipLabel))
		emitter.emitLabel(tokenName)
		word2addr[tokenName] = emitter.curAddr
		funcSkips = append(funcSkips, skipLabel)
	case ";":
		if !isCurrentUnsafe {
			if currentDepth != expectedPuts {
				return index, fmt.Errorf("contract mismatch in %s procedure: expected %d elements, got %d", currentProc, expectedPuts, currentDepth)
			}
		}
		currentDepth = savedDepth

		emitter.emitRet()
		if len(funcSkips) == 0 {
			return index, fmt.Errorf("syntax error: ':' expected")
		}
		last := len(funcSkips) - 1
		skipLabel := funcSkips[last]
		funcSkips = funcSkips[:last]
		emitter.emitLabel(skipLabel)
	case "'":
		index++
		nextToken := tokens[index]
		tokenName := strings.ToUpper(nextToken.valStr)
		emitter.emitLit(tokenName)
		currentDepth++
	case "CELLS":
		emitter.emitLoadImmNum(isa.T2, 4)
		emitter.emit(isa.MUL(isa.T0, isa.T0, isa.T2))
	case "EXECUTE":
		emitter.emit(isa.MV(isa.T2, isa.T0))
		emitter.emit(isa.MV(isa.T0, isa.T1))
		emitter.emit(isa.LW(isa.T1, isa.SP, 0))
		emitter.emit(isa.ADDI(isa.SP, isa.SP, 4))

		retLabel := fmt.Sprintf("_EXEC_RET_%d", emitter.curAddr)
		emitter.emitLoadImmLabel(isa.T3, retLabel)
		emitter.emit(isa.ADDI(isa.RP, isa.RP, -4))
		emitter.emit(isa.SW(isa.T3, isa.RP, 0))
		emitter.emit(isa.JR(isa.T2))
		emitter.emitLabel(retLabel)
		currentDepth--
	case "VAR":
		index++
		nextToken := tokens[index]
		varName := strings.ToUpper(nextToken.valStr)

		if !assertFreeName(varName) {
			return index, fmt.Errorf("name %s already defined", varName)
		}

		varLabel := fmt.Sprintf("VAR_%s", varName)
		var2Label[varName] = varLabel
		emitter.emitData(isa.Data{Value: 0, ProgramItem: isa.ProgramItem{Label: varLabel}})
	case "ARRAY":
		index++
		nextToken := tokens[index]
		arrName := strings.ToUpper(nextToken.valStr)

		if !assertFreeName(arrName) {
			return index, fmt.Errorf("name %s already defined", arrName)
		}
		emitter.emit(isa.MV(isa.T0, isa.T1))
		emitter.emit(isa.LW(isa.T1, isa.SP, 0))
		emitter.emit(isa.ADDI(isa.SP, isa.SP, 4))

		arrLabel := fmt.Sprintf("ARRAY_%s", arrName)
		var2Label[arrName] = arrLabel
		if lastNumber > 0 {
			emitter.emitData(isa.Data{Value: 0, ProgramItem: isa.ProgramItem{Label: arrLabel}})
			for range lastNumber - 1 {
				emitter.emitData(isa.Data{Value: 0})
			}
		}
	case "STRING":
		index++
		nextToken := tokens[index]
		strVal := nextToken.valStr
		index++
		nextToken = tokens[index]
		strName := nextToken.valStr

		if !assertFreeName(strName) {
			return index, fmt.Errorf("name %s already defined", strName)
		}
		label := fmt.Sprintf("STRING_%s", strName)
		var2Label[strName] = label
		emitter.emitData(isa.Data{Value: int32(utf8.RuneCountInString(strVal)), ProgramItem: isa.ProgramItem{Label: label}})

		for _, char := range strVal {
			emitter.emitData(isa.Data{Value: char})
		}
	default:
		if contract, ok := funcContracts[word]; ok {
			if currentDepth < contract.Takes {
				return index, fmt.Errorf("stack underflow calling %s", word)
			}
			currentDepth -= contract.Takes
			currentDepth += contract.Puts

			emitter.emitCall(strings.ToUpper(word))
		} else if label, ok := var2Label[word]; ok {
			emitter.emitLit(label)
			currentDepth++
		} else {
			return index, fmt.Errorf("unknown word %s", word)
		}
	}
	return index, nil
}

func assertFreeName(name string) bool {
	if _, ok := var2Label[name]; !ok {
		return true
	}
	if _, ok := word2addr[name]; !ok {
		return true
	}
	return false
}

func peephole(tokens []token, index int, emitter *Emitter) (bool, int) {
	ops := map[string]string{
		"+":   "add",
		"-":   "sub",
		"*":   "mul",
		"/":   "div",
		"DIV": "div",
		"%":   "mod",
		"MOD": "mod",
		"AND": "and",
	}
	token := tokens[index]

	// <num> <operation>
	if token.typ == "NUMBER" && index+1 < len(tokens) {
		nextToken := tokens[index+1]
		if op, ok := ops[nextToken.valStr]; ok {
			val := token.valNum

			if (op == "add" || op == "sub") && val >= -2048 && val <= 2047 {
				if op == "add" {
					emitter.emit(isa.ADDI(isa.T0, isa.T0, val))
				} else {
					emitter.emit(isa.ADDI(isa.T0, isa.T0, -val))
				}
			} else {
				emitter.emitLoadImmNum(isa.T2, val)
				switch op {
				case "add":
					emitter.emit(isa.ADD(isa.T0, isa.T0, isa.T2))
				case "sub":
					emitter.emit(isa.SUB(isa.T0, isa.T0, isa.T2))
				case "mul":
					emitter.emit(isa.MUL(isa.T0, isa.T0, isa.T2))
				case "div":
					emitter.emit(isa.DIV(isa.T0, isa.T0, isa.T2))
				case "mod":
					emitter.emit(isa.MOD(isa.T0, isa.T0, isa.T2))
				case "and":
					emitter.emit(isa.AND(isa.T0, isa.T0, isa.T2))
				}
			}
			newIndex := index + 2
			return true, newIndex
		}
		// <var> @
	} else if token.typ == "WORD" &&
		index+1 < len(tokens) &&
		tokens[index+1].valStr == "@" {
		if _, ok := var2Label[token.valStr]; ok {
			label := var2Label[token.valStr]
			emitter.emit(isa.ADDI(isa.SP, isa.SP, -4))
			emitter.emit(isa.SW(isa.T1, isa.SP, 0))
			emitter.emit(isa.MV(isa.T1, isa.T0))
			emitter.emit(isa.LUIHi(isa.T0, label))
			emitter.emit(isa.ADDILo(isa.T0, isa.T0, label))
			emitter.emit(isa.LW(isa.T0, isa.T0, 0))

			currentDepth++
			newIndex := index + 2
			return true, newIndex
		}
		// <var> !
	} else if token.typ == "WORD" &&
		index+1 < len(tokens) &&
		tokens[index+1].valStr == "!" {
		if _, ok := var2Label[token.valStr]; ok {
			label := var2Label[token.valStr]
			emitter.emit(isa.LUIHi(isa.T2, label))
			emitter.emit(isa.ADDILo(isa.T2, isa.T2, label))
			emitter.emit(isa.SW(isa.T0, isa.T2, 0))

			emitter.emit(isa.MV(isa.T0, isa.T1))
			emitter.emit(isa.LW(isa.T1, isa.SP, 0))
			emitter.emit(isa.ADDI(isa.SP, isa.SP, 4))

			currentDepth--
			newIndex := index + 2
			return true, newIndex
		}
	}
	return false, index
}

func resolveAddresses(program isa.Program) isa.Program {
	instructionCnt := len(program.Instructions)
	curAddr := program.Instructions[instructionCnt-1].Address + wordSize

	for i := range program.Data {
		program.Data[i].Address = curAddr
		curAddr += wordSize
	}
	return program
}

func resolveLabels(program isa.Program) (isa.Program, error) {
	label2addr := make(map[string]int)

	for _, instr := range program.Instructions {
		if instr.Label != "" {
			labels := strings.Split(instr.Label, ":")
			for _, l := range labels {
				label2addr[l] = instr.Address
			}
		}
	}

	for _, data := range program.Data {
		if data.Label != "" {
			labels := strings.Split(data.Label, ":")
			for _, l := range labels {
				label2addr[l] = data.Address
			}
		}
	}

	for i := range program.Instructions {
		instr := &program.Instructions[i]

		if instr.TargetLabel != "" {
			label := instr.TargetLabel
			imm, ok := label2addr[label]
			if !ok {
				return program, fmt.Errorf("label %s not found", label)
			}

			val := int32(imm)
			switch instr.Macros {
			case isa.MacroHi:
				upper := (val >> 12) & 0xFFFFF
				lower := val & 0xFFF
				if lower&0x800 != 0 {
					upper++
				}
				program.Instructions[i].Imm = upper
			case isa.MacroLo:
				program.Instructions[i].Imm = val & 0xFFF

			default:
				program.Instructions[i].Imm = val
			}
		}
	}

	return program, nil
}

func Translate(sourcePath, targetPath string) {
	bytes, err := os.ReadFile(sourcePath)
	if err != nil {
		log.Fatal(err)
	}
	sourceCode := string(bytes)
	tokens, err := tokenize(sourceCode)
	if err != nil {
		log.Fatal(err)
	}

	var libs []string
	var userTokens []token
	for _, token := range tokens {
		if token.typ == "IMPORT" {
			libs = append(libs, token.valStr)
		} else {
			userTokens = append(userTokens, token)
		}
	}

	var libsText []string
	for _, lib := range libs {
		libPath := fmt.Sprintf("libs/%s.fth", lib)
		bytes, err = os.ReadFile(libPath)
		if err != nil {
			log.Fatal(err)
		}
		libsText = append(libsText, string(bytes))
	}
	libsTokens, err := tokenize(strings.Join(libsText, "\n"))
	if err != nil {
		log.Fatal(err)
	}

	emitter := NewEmitter()
	emitter.emitLoadImmNum(isa.SP, dataStackInitAddr)
	emitter.emitLoadImmNum(isa.RP, returnStackInitAddr)

	err = translateTokens(libsTokens, emitter)
	if err != nil {
		log.Fatal(err)
	}

	libInstCount := len(emitter.program.Instructions)
	libDataCount := len(emitter.program.Data)

	err = translateTokens(userTokens, emitter)
	if err != nil {
		log.Fatal(err)
	}
	emitter.emit(isa.HALT())

	rawProgram := emitter.program
	program := resolveAddresses(rawProgram)

	program, err = resolveLabels(program)
	if err != nil {
		log.Fatal(err)
	}

	userProgram := isa.Program{
		Instructions: program.Instructions[libInstCount:],
		Data:         program.Data[libDataCount:],
	}

	hexCode := isa.ToHex(userProgram)
	hexData := strings.Join(hexCode, "\n")

	binaryCode := isa.ToBytes(program)

	err = os.WriteFile(targetPath+".hex", []byte(hexData), os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(targetPath, binaryCode, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
}
