package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
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
	directory := subParser.String("d", getDefaultReminderDir(), "Directory under which to create .daily-reminder")

	subParser.Parse(args[1:])

	dr := getReminder(*directory, true)
	dr.Init(true)
}

func handleStatusCmd(args []string) {
	drDir := getDefaultReminderDir()
	fmt.Printf("drDir=%s\n", drDir)
	dr := getReminder(drDir, false)
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

	dr := getReminder(getDefaultReminderDir(), false)
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

	from, err := parseCmdTime(*fromStr)
	if err != nil {
		fmt.Printf("Failed to parse from date (%s): %v\n", *fromStr, err)
		return
	}

	to, err := parseCmdTime(*toStr)
	if err != nil {
		fmt.Printf("Failed to parse to date (%s): %v\n", *toStr, err)
		return
	}

	mode := reminder.ModeRegular
	if *expand {
		mode = reminder.ModeExpand
	}

	dr := getReminder(getDefaultReminderDir(), false)
	dr.Init(true)

	results := dr.Query(reminder.QueryParam{
		From:   from,
		To:     to,
		Traits: traits,
		Mode:   mode,
	})

	if len(results) == 0 {
		fmt.Printf("(Empty)\n")
		return
	}

	// sort by time
	slices.SortFunc(results, func(r1 reminder.QueryResult, r2 reminder.QueryResult) int {
		return time.Time.Compare(r1.Time, r2.Time)
	})

	for _, r := range results {
		fmt.Printf("%s:%s;%s\n", r.Date.ToString(), r.Record.Name, strings.Join(r.Record.Traits, ","))
	}
}

func getReminder(path string, isForce bool) reminder.Reminder {
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

func getDefaultReminderDir() string {
	// first we check env
	envReminderPath := os.Getenv("DR_DIR")
	if envReminderPath != "" {
		return envReminderPath
	}

	// if path in env not available,
	// we use $HOME as default path
	return os.Getenv("HOME")
}

func parseCmdTime(timeStr string) (tm time.Time, err error) {
	// 1. fancy time
	if tm, isFancy := parseFancyDate(timeStr); isFancy {
		return tm, nil
	}

	// 2. normal date
	tm, err = time.Parse("2006/01/02", timeStr)
	return
}

func parseFancyDate(dateStr string) (time.Time, bool) {
	ts := strings.ToLower(strings.TrimSpace(dateStr))

	// date literal
	switch ts {
	case "today":
		return reminder.StartOfDay(time.Now()), true
	case "yesterday":
		return reminder.StartOfDay(time.Now()).AddDate(0, 0, -1), true
	case "tomorrow":
		return reminder.StartOfDay(time.Now()).AddDate(0, 0, 1), true
	case "":
		return time.Time{}, true
	}

	// date arithmetic
	var delta int
	if cnt, err := fmt.Sscanf(ts, "%d days ago", &delta); cnt == 1 && err == nil {
		return reminder.StartOfDay(time.Now()).AddDate(0, 0, -delta), true
	}
	if cnt, err := fmt.Sscanf(ts, "%d days later", &delta); cnt == 1 && err == nil {
		return reminder.StartOfDay(time.Now()).AddDate(0, 0, delta), true
	}
	if cnt, err := fmt.Sscanf(ts, "%d months ago", &delta); cnt == 1 && err == nil {
		return reminder.StartOfDay(time.Now()).AddDate(0, -delta, 0), true
	}
	if cnt, err := fmt.Sscanf(ts, "%d months later", &delta); cnt == 1 && err == nil {
		return reminder.StartOfDay(time.Now()).AddDate(0, delta, 0), true
	}
	if cnt, err := fmt.Sscanf(ts, "%d years ago", &delta); cnt == 1 && err == nil {
		return reminder.StartOfDay(time.Now()).AddDate(-delta, 0, 0), true
	}
	if cnt, err := fmt.Sscanf(ts, "%d years later", &delta); cnt == 1 && err == nil {
		return reminder.StartOfDay(time.Now()).AddDate(delta, 0, 0), true
	}

	return time.Time{}, false
}
