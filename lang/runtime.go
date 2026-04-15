package lang

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/Alexzjc2003/daily-reminder/reminder"
)

var r reminder.Reminder

func GetReminder() reminder.Reminder { return r }

type VariableTable map[string]Variable

var vt VariableTable

func GetVT() VariableTable { return vt }

type CmdRunner func([]string) (Variable, error)

var runners map[string]CmdRunner

func InitRuntime() {
	// 1. get reminder

	var path string

	envReminderPath := os.Getenv("DR_DIR")
	if envReminderPath != "" {
		path = envReminderPath
	}

	// if path in env not available,
	// we use $HOME as default path
	path = os.Getenv("HOME")

	if !reminder.IsDir(path) {
		panic(fmt.Errorf("%s: Not a directory\n", path))
	}
	drPath := filepath.Join(path, ".daily-reminder")

	if reminder.IsExist(drPath) {
		if !reminder.IsDir(drPath) {
			panic(fmt.Errorf("%s: Already exists but not a directory\n", drPath))
		}
	} else {
		panic(fmt.Errorf("%s: Does not exist\n", drPath))
	}

	r := reminder.Reminder{
		Path: drPath,
	}

	r.Init(false)

	// 2. init variable table
	vt = map[string]Variable{}

	// 3. init cmd runners
	runners = map[string]CmdRunner{
		"remember": nil,
		"query":    queryRunner,
		"typeof":   typeofRunner,
	}
}

func IsCmd(cmd string) bool {
	_, ok := runners[cmd]
	return ok
}

func RunCmd(cmd string, args []string) (Variable, error) {
	if runner, ok := runners[cmd]; ok {
		return runner(args)
	} else {
		return VariableEmpty{}, fmt.Errorf("not a cmd")
	}
}

func typeofRunner(args []string) (Variable, error) {
	if len(args) < 1 {
		return VariableEmpty{}, fmt.Errorf("not enough args for kindof")
	}

	expr, err := ParseExpr(args)
	if err != nil {
		return VariableEmpty{}, err
	}

	switch expr.(type) {
	case VariableDate:
		return VariableString{Data: "Date"}, nil
	case VariableEmpty:
		return VariableString{Data: "Empty"}, nil
	case VariableList:
		return VariableString{Data: "List"}, nil
	case VariableNumber:
		return VariableString{Data: "Number"}, nil
	case VariableString:
		return VariableString{Data: "String"}, nil
	default:
		return VariableString{Data: "Unknown"}, nil
	}
}

func queryRunner(args []string) (Variable, error) {
	traits := []string{}
	subParser := flag.NewFlagSet("query", flag.ExitOnError)
	fromStr := subParser.String("from", "", "From date (YYYY/MM/DD)")
	toStr := subParser.String("to", "", "To date (YYYY/MM/DD)")
	expand := subParser.Bool("x", false, "Expand dates")
	subParser.Func("t", "Specify traits, both -t t1 -t t2 and -t 1,2 can be accepted", func(traitStr string) error {
		for trait := range strings.SplitSeq(traitStr, ",") {
			if strings.TrimSpace(trait) == "" {
				continue
			}
			traits = append(traits, trait)
		}
		return nil
	})

	err := subParser.Parse(args)
	if err != nil {
		return VariableEmpty{}, err
	}

	var queryParam reminder.QueryParam

	if v, err := ParseExpr([]string{*fromStr}); err != nil {
		return VariableEmpty{}, err
	} else {
		switch d := v.(type) {
		case VariableDate:
			queryParam.From = d.Data.ToTime()
		case VariableString:
			// maybe fancy
			if fd, ok := reminder.ParseFancyDate(d.Data); !ok {
				return VariableEmpty{}, fmt.Errorf("illegal from date: non-fancy string %v", d.ToString())
			} else {
				queryParam.From = fd.ToTime()
			}
		default:
			return VariableEmpty{}, fmt.Errorf("illegal from date: wrong type")
		}
	}

	if to, err := ParseExpr([]string{*toStr}); err != nil {
		return VariableEmpty{}, err
	} else {
		switch d := to.(type) {
		case VariableDate:
			queryParam.To = d.Data.ToTime()
		case VariableString:
			// maybe fancy
			if fd, ok := reminder.ParseFancyDate(d.Data); !ok {
				return VariableEmpty{}, fmt.Errorf("illegal to date: non-fancy string (%v)", d.ToString())
			} else {
				queryParam.To = fd.ToTime()
			}
		default:
			return VariableEmpty{}, fmt.Errorf("illegal to date: wrong type (%v)", d.ToString())
		}
	}

	if *expand {
		queryParam.Mode = reminder.ModeExpand
	} else {
		queryParam.Mode = reminder.ModeRegular
	}

	dr := GetReminder()

	results := dr.Query(queryParam)

	// sort by time
	slices.SortFunc(results, func(r1 reminder.QueryResult, r2 reminder.QueryResult) int {
		return time.Time.Compare(r1.Time, r2.Time)
	})

	res := VariableList{}
	for _, r := range results {
		// fmt.Printf("%s:%s;%s\n", r.Date.ToString(), r.Record.Name, strings.Join(r.Record.Traits, ","))
		res.Data = append(res.Data, VariableString{Data: fmt.Sprintf("%s:%s;%s\n", r.Date.ToString(), r.Record.Name, strings.Join(r.Record.Traits, ","))})
	}

	return res, nil
}
