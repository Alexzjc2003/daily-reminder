package lang

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/google/shlex"

	"github.com/Alexzjc2003/daily-reminder/reminder"
)

func RunFile(reminder reminder.Reminder, filename string) error {
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

	// try parse as a variable
	if v, ok := strings.CutPrefix(args[0], "$"); ok {
		// Note this is not cfg, that variable parsing should
		// succeed only if it is already defined.
		if data, ok := GetVT()[v]; ok {
			return data, nil
		}
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
