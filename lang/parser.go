package lang

import (
	"bytes"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/google/shlex"

	"github.com/Alexzjc2003/daily-reminder/reminder"
)

func RunFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	buf := bytes.Buffer{}
	buf.ReadFrom(file)

	content := strings.TrimSpace(buf.String())
	if len(content) == 0 {
		return fmt.Errorf("Empty file")
	}

	lines := splitAndNormalizeLine(content)

	for _, line := range lines {
		// support comments
		if strings.HasPrefix(line, "//") {
			continue
		}

		args, err := shlex.Split(line)
		if err != nil {
			return err
		}

		if err := ParseCmd(args); err != nil {
			return err
		}
	}

	return nil
}

func ParseCmd(args []string) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("empty command")
	}

	cmd := args[0]

	switch cmd {
	case "set":
		err = ParseSetCmd(args)
	case "print":
		err = ParsePrintCmd(args)
	case "foreach":
		err = ParseForeachCmd(args)
	default:
		return fmt.Errorf("unknown command: %v", cmd)
	}

	return
}

func ParseSetCmd(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("not enough args for set")
	}

	// expect a variable
	v, ok := strings.CutPrefix(args[1], "$")
	if !ok {
		return fmt.Errorf("expected a var, got %v, maybe $%v?", args[1], args[1])
	}

	expr, err := ParseExpr(args[2:])
	if err != nil {
		return err
	}

	GetVT()[v] = expr

	return nil
}

func ParsePrintCmd(args []string) error {
	v, err := ParseExpr(args[1:])
	if err != nil {
		return err
	}

	fmt.Printf("%v\n", v.ToString())
	return nil
}

func ParseForeachCmd(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("wrong foreach format")
	}

	// find do, which marks the start of foreach body
	doArgIndex := slices.Index(args, "do")
	if doArgIndex < 0 {
		return fmt.Errorf("wrong foreach format: missing do")
	}

	list, err := ParseExpr(args[1:doArgIndex])
	if err != nil {
		return err
	}

	vt := GetVT()
	// for now we are only supporting one-line body
	// we'll impl the rest later
	switch v := list.(type) {
	case VariableList:
		for i, e := range v.Data {
			// TODO: probably we can maintain a stack here for nested loops
			// store loop var
			vt["%index"] = VariableNumber{Data: int64(i)}
			vt["%value"] = e
			// start looping
			if err := ParseCmd(args[doArgIndex+1:]); err != nil {
				return err
			}
		}
		delete(vt, "%index")
		delete(vt, "%value")
	case VariableObject:
		for k, v := range v.Data {
			// TODO: probably we can maintain a stack here for nested loops
			// store loop var
			vt["%key"] = VariableString{Data: k}
			vt["%value"] = v
			// start looping
			if err := ParseCmd(args[doArgIndex+1:]); err != nil {
				return err
			}
		}
		delete(vt, "%key")
		delete(vt, "%value")
	default:
		return fmt.Errorf("foreach: target variable not iterable")
	}

	return nil
}

func ParseExpr(args []string) (Variable, error) {
	if len(args) < 1 {
		return VariableString{}, fmt.Errorf("empty expression")
	}

	// try parse as a number
	if data, err := strconv.ParseInt(args[0], 10, 64); err == nil {
		return VariableNumber{Data: data}, nil
	}

	// try parse as a date
	if date, ok := strings.CutPrefix(args[0], "@"); ok {
		if data, err := reminder.ParseDate(date); err == nil {
			return VariableDate{Data: data}, nil
		}
	}

	// try parse as a special variable
	if strings.HasPrefix(args[0], "%") {
		segs := strings.Split(args[0], ".")

		if _, ok := GetVT()[segs[0]]; !ok {
			return VariableEmpty{}, fmt.Errorf("fetching special variable(%s) out of scope", segs[0])
		} else {
			return ParseFieldAccess(segs)
		}
	}

	// try parse as a standard variable
	if v, ok := strings.CutPrefix(args[0], "$"); ok {
		// Support field access
		segs := strings.Split(v, ".")

		return ParseFieldAccess(segs)
	}

	// try parse as cmd
	if IsCmd(args[0]) {
		return RunCmd(args[0], args[1:])
	}

	return VariableString{Data: strings.Join(args, " ")}, nil
}

func splitAndNormalizeLine(s string) (result []string) {
	lines := strings.Split(s, "\n")

	var current string
	for _, line := range lines {
		line = strings.TrimRight(line, "\r")

		if before, ok := strings.CutSuffix(line, "\\"); ok {
			current += before
		} else {
			current += line
			if trimmed := strings.TrimSpace(current); trimmed != "" {
				result = append(result, trimmed)
			}
			current = ""
		}
	}

	if current != "" {
		result = append(result, strings.TrimSpace(current))
	}

	return
}

func ParseFieldAccess(segs []string) (Variable, error) {
	var va Variable
	for idx, seg := range segs {
		if idx == 0 {
			if val, ok := GetVT()[seg]; !ok {
				return VariableEmpty{}, fmt.Errorf("variable not defined(%s)", seg)
			} else {
				va = val
				continue
			}
		}

		if vl, ok := va.(VariableList); ok {
			// expect seg to be a Number
			if index, err := strconv.ParseInt(seg, 10, 64); err != nil {
				return VariableEmpty{}, fmt.Errorf("expect index for list(%s), got %s", strings.Join(segs[:idx], "."), seg)
			} else {
				va = vl.Data[index]
				continue
			}
		}

		if vo, ok := va.(VariableObject); ok {
			if val, ok := vo.Data[seg]; !ok {
				return VariableEmpty{}, fmt.Errorf("field %s does not exist on object(%s)", seg, strings.Join(segs[:idx], "."))
			} else {
				va = val
				continue
			}
		}
	}

	return va, nil
}
