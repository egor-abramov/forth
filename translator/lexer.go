package translator

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type token struct {
	typ    string
	valNum int32
	valStr string
}

func tokenize(code string) ([]token, error) {
	var tokens []token

	tokenSpecification := []struct {
		name    string
		pattern string
	}{
		{"IMPORT", `import\s+(\w+)`},
		{"STRING", `"[^"]*"`},
		{"NUMBER", `-?\d+`},
		{"WORD", `[^\s]+`},
	}

	var regexParts []string
	for _, spec := range tokenSpecification {
		regexParts = append(regexParts, fmt.Sprintf("(?P<%s>%s)", spec.name, spec.pattern))
	}
	tokenRegexp := strings.Join(regexParts, "|")
	re := regexp.MustCompile(tokenRegexp)

	matches := re.FindAllStringSubmatch(code, -1)
	groupNames := re.SubexpNames()

	for _, match := range matches {
		var typ, val string

		for i, name := range groupNames {
			if i != 0 && name != "" && match[i] != "" {
				typ = name
				val = match[0]
				break
			}
		}

		switch typ {
		case "NUMBER":
			parsedNum, _ := strconv.ParseInt(val, 10, 32)
			tokens = append(tokens, token{typ: typ, valNum: int32(parsedNum)})
		case "WORD":
			tokens = append(tokens, token{typ: typ, valStr: strings.ToUpper(val)})
		case "STRING":
			actualStr := val[1 : len(val)-1]
			unquoted, err := strconv.Unquote(`"` + actualStr + `"`)
			if err == nil {
				actualStr = unquoted
			}
			tokens = append(tokens, token{typ: typ, valStr: actualStr})
		case "IMPORT":
			parts := strings.Fields(val)
			if len(parts) > 1 {
				libName := strings.ToLower(parts[1])
				tokens = append(tokens, token{typ: typ, valStr: libName})
			}
		}
	}

	return tokens, nil
}
