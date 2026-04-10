package reminder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Reminder struct {
	Path  string
	Dates []ReminderRecord
}

func (r *Reminder) Init(isForce bool) {
	datesPath := filepath.Join(r.Path, "dates")
	if !IsExist(datesPath) {
		if isForce {
			initDate := time.Now().Format("2006/01/02:DR INIT;\n")
			os.WriteFile(datesPath, []byte(initDate), os.ModePerm)
		} else {
			panic(fmt.Errorf("Dates file does not exist: %s\n", datesPath))
		}
	}

	datesRaw, err := os.ReadFile(datesPath)
	if err != nil {
		panic(fmt.Errorf("Failed to read %s: %v\n", datesPath, err))
	}

	for dateString := range strings.SplitSeq(string(datesRaw), "\n") {
		if strings.TrimSpace(dateString) == "" {
			continue
		}

		var date ReminderRecord
		if err := date.FromString(dateString); err != nil {
			fmt.Printf("Error parsing date (%s): %v\n", dateString, err)
		} else {
			r.Dates = append(r.Dates, date)
		}
	}
}

func (r *Reminder) Remember(date ReminderRecord, shouldWrite bool) {
	r.Dates = append(r.Dates, date)
	fmt.Println(date.ToString())
	if !shouldWrite {
		return
	}

	datesFile, err := os.OpenFile(filepath.Join(r.Path, "dates"), os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		panic(fmt.Errorf("Failed to open dates file:\n%v", err))
	}
	defer datesFile.Close()

	datesFile.WriteString(date.ToString())
	datesFile.WriteString("\n")
}

func (r Reminder) Query(query QueryParam) []QueryResult {
	var q ReminderQuery
	q.Build(query)

	return q.Apply(r)
}
