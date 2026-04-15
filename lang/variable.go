package lang

import (
	"strconv"
	"strings"

	"github.com/Alexzjc2003/daily-reminder/reminder"
)

type Variable interface {
	ToString() string
}

type VariableString struct {
	Data string
}

type VariableNumber struct {
	Data int64
}

type VariableDate struct {
	Data reminder.ReminderDate
}

type VariableList struct {
	Data []Variable
}

type VariableEmpty struct{}

func (vn VariableNumber) ToString() string {
	return strconv.FormatInt(vn.Data, 10)
}

func (vs VariableString) ToString() string {
	return vs.Data
}

func (vd VariableDate) ToString() string {
	return vd.Data.ToString()
}

func (vl VariableList) ToString() string {
	var vs []string
	for _, v := range vl.Data {
		vs = append(vs, v.ToString())
	}

	return "[" + strings.Join(vs, ",") + "]"
}

func (ve VariableEmpty) ToString() string {
	return "(empty)"
}
