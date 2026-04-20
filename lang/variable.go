package lang

import (
	"fmt"
	"reflect"
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

type VariableObject struct {
	Data map[string]Variable
}

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

	return "[ " + strings.Join(vs, ", ") + " ]"
}

func (ve VariableEmpty) ToString() string {
	return "(empty)"
}

func (vo VariableObject) ToString() string {
	var kvp []string
	for k, v := range vo.Data {
		kvp = append(kvp, fmt.Sprintf("%s: %s", k, v.ToString()))
	}

	return fmt.Sprintf("{ %s }", strings.Join(kvp, ", "))
}

func (vo *VariableObject) FromMap(m map[string]any) {
	for k, v := range m {
		if variable, err := FromReflect(reflect.ValueOf(v)); err != nil {
			continue
		} else {
			vo.Data[k] = variable
		}
	}
}

func FromReflect(val reflect.Value) (Variable, error) {
	switch val.Kind() {
	case reflect.Pointer, reflect.Interface:
		return FromReflect(val.Elem())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return VariableNumber{Data: val.Int()}, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return VariableNumber{Data: int64(val.Uint())}, nil
	case reflect.Array, reflect.Slice:
		var l []Variable
		for i := 0; i < val.Len(); i++ {
			vi, _ := FromReflect(val.Index(i))
			l = append(l, vi)
		}
		return VariableList{Data: l}, nil
	case reflect.Map:
		m := map[string]Variable{}
		iter := val.MapRange()
		for iter.Next() {
			k, v := iter.Key(), iter.Value()
			// we are somehow assuming here that k is a string
			vv, _ := FromReflect(v)
			m[k.String()] = vv
		}
		return VariableObject{Data: m}, nil
	case reflect.String:
		return VariableString{Data: val.String()}, nil
	default:
		return VariableEmpty{}, fmt.Errorf("can not parse from go object")
	}
}
