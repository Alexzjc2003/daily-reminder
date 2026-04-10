package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"./reminder"
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("See dr help")
		return
	}

	cmd := os.Args[1]

	switch cmd {
	case "help":
		handleHelpCmd(os.Args[1:])
	case "init":
		handleInitCmd(os.Args[1:])
	case "status":
		handleStatusCmd(os.Args[1:])
	case "remember":
		handleRememberCmd(os.Args[1:])
	case "query":
		handleQueryCmd(os.Args[1:])
	default:
		handleHelpCmd(os.Args[1:])
	}

}

func handleHelpCmd(args []string) {
	fmt.Printf("Help message for dr\n")
}

func handleInitCmd(args []string) {
	subParser := flag.NewFlagSet("init", flag.ExitOnError)

	// isForce := subParser.Bool("f", false, "Force create .daily-reminder/")
	directory := subParser.String("d", GetDefaultReminderDir(), "Directory under which to create .daily-reminder")

	subParser.Parse(args[1:])

	dr := GetReminder(*directory, true)
	dr.Init(true)
}

func handleStatusCmd(args []string) {
	drDir := GetDefaultReminderDir()
	fmt.Printf("drDir=%s\n", drDir)
	dr := GetReminder(drDir, false)
	dr.Init(true)

	for _, d := range dr.Dates {
		fmt.Println(d.ToString())
	}
}

func handleRememberCmd(args []string) {
	traits := []string{}
	subParser := flag.NewFlagSet("remember", flag.ExitOnError)
	subParser.Func("t", "Specify traits, both -t t1 -t t2 and -t 1,2 can be accepted", func(traitStr string) error {
		for trait := range strings.SplitSeq(traitStr, ",") {
			if strings.TrimSpace(trait) == "" {
				continue
			}
			traits = append(traits, trait)
		}

		return nil
	})
	readonly := subParser.Bool("r", false, "Readonly, don't write into dates file")

	err := subParser.Parse(args[1:])
	if err != nil {
		println(err)
	}

	if subParser.NArg() < 2 {
		fmt.Printf("Usage: dr remember [-r] [-t value] <name> <date>\n")
		subParser.Usage()
		return
	}

	name := subParser.Arg(0)
	dateStr := subParser.Arg(1)

	dr := GetReminder(GetDefaultReminderDir(), false)
	dr.Init(true)

	date, err := reminder.ParseDate(dateStr)
	if err != nil {
		fmt.Printf("Failed to parse (%s): %v\n", dateStr, err)
		return
	}

	dr.Remember(reminder.ReminderRecord{
		Name:   name,
		Date:   date,
		Traits: traits,
	}, !*readonly)
}

func handleQueryCmd(args []string) {
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

	err := subParser.Parse(args[1:])
	if err != nil {
		println(err)
		return
	}

	var from *time.Time
	var to *time.Time
	if *fromStr != "" {
		tm, err := time.Parse("2006/01/02", *fromStr)
		if err != nil {
			fmt.Printf("Failed to parse from date (%s): %v\n", *fromStr, err)
			return
		}
		from = &tm
	}
	if *toStr != "" {
		tm, err := time.Parse("2006/01/02", *toStr)
		if err != nil {
			fmt.Printf("Failed to parse to date (%s): %v\n", *toStr, err)
			return
		}
		to = &tm
	}

	mode := reminder.ModeRegular
	if *expand {
		mode = reminder.ModeExpand
	}

	dr := GetReminder(GetDefaultReminderDir(), false)
	dr.Init(true)

	results := dr.Query(reminder.QueryParam{
		From:   from,
		To:     to,
		Traits: traits,
		Mode:   mode,
	})

	for _, r := range results {
		fmt.Printf("%s:%s;%s\n", r.Date.ToString(), r.Record.Name, strings.Join(r.Record.Traits, ","))
	}
}

func GetReminder(path string, isForce bool) reminder.Reminder {
	if !reminder.IsDir(path) {
		panic(fmt.Errorf("%s: Not a directory\n", path))
	}
	drPath := filepath.Join(path, ".daily-reminder")

	if reminder.IsExist(drPath) {
		if !reminder.IsDir(drPath) {
			panic(fmt.Errorf("%s: Already exists but not a directory\n", drPath))
		}
	} else {
		if isForce {
			os.Mkdir(drPath, os.ModePerm)
		} else {
			panic(fmt.Errorf("%s: Does not exist\n", drPath))
		}
	}

	return reminder.Reminder{
		Path: drPath,
	}
}

func GetDefaultReminderDir() string {
	// first we check env
	envReminderPath := os.Getenv("DR_DIR")
	if envReminderPath != "" {
		return envReminderPath
	}

	// if path in env not available,
	// we use $HOME as default path
	return os.Getenv("HOME")
}
