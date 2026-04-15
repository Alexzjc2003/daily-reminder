package reminder

import (
	"fmt"
	"strings"
	"time"
)

type ReminderDate struct {
	Year  uint
	Month uint
	Day   uint
}

type DateKind uint

const (
	NODATE DateKind = iota
	DAYONLY
	MONTHONLY
	MONTHDAY
	YEARONLY
	YEARDAY
	YEARMONTH
	FULLDATE
)

func (date ReminderDate) Kind() DateKind {
	kind := 0
	if date.Year != 0 {
		kind += 4
	}
	if date.Month != 0 {
		kind += 2
	}
	if date.Day != 0 {
		kind += 1
	}
	return DateKind(kind)
}

func ParseDate(dateStr string) (date ReminderDate, err error) {
	_, err = fmt.Sscanf(dateStr, "%d/%d/%d", &date.Year, &date.Month, &date.Day)
	return
}

func (d ReminderDate) ToString() string {
	return fmt.Sprintf("%04d/%02d/%02d", d.Year, d.Month, d.Day)
}

func (d ReminderDate) ToTime() time.Time {
	if d.Kind() == FULLDATE {
		return time.Date(int(d.Year), time.Month(d.Month), int(d.Day), 0, 0, 0, 0, time.Local)
	} else {
		return time.Time{}
	}
}

func (d *ReminderDate) FromTime(tm time.Time) {
	d.Year = uint(tm.Year())
	d.Month = uint(tm.Month())
	d.Day = uint(tm.Day())
}

func ParseFancyDate(dateStr string) (date ReminderDate, isFancy bool) {
	isFancy = true
	ts := strings.ToLower(strings.TrimSpace(dateStr))

	// date literal
	switch ts {
	case "today":
		date.FromTime(time.Now())
		return
	case "yesterday":
		date.FromTime(time.Now().AddDate(0, 0, -1))
		return
	case "tomorrow":
		date.FromTime(time.Now().AddDate(0, 0, 1))
		return
	case "":
		date.FromTime(time.Time{})
		return
	}

	// date arithmetic
	var delta int
	if cnt, err := fmt.Sscanf(ts, "%d days ago", &delta); cnt == 1 && err == nil {
		date.FromTime(time.Now().AddDate(0, 0, -delta))
		return
	}
	if cnt, err := fmt.Sscanf(ts, "%d days later", &delta); cnt == 1 && err == nil {
		date.FromTime(time.Now().AddDate(0, 0, delta))
		return
	}
	if cnt, err := fmt.Sscanf(ts, "%d months ago", &delta); cnt == 1 && err == nil {
		date.FromTime(time.Now().AddDate(0, -delta, 0))
		return
	}
	if cnt, err := fmt.Sscanf(ts, "%d months later", &delta); cnt == 1 && err == nil {
		date.FromTime(time.Now().AddDate(0, delta, 0))
		return
	}
	if cnt, err := fmt.Sscanf(ts, "%d years ago", &delta); cnt == 1 && err == nil {
		date.FromTime(time.Now().AddDate(-delta, 0, 0))
		return
	}
	if cnt, err := fmt.Sscanf(ts, "%d years later", &delta); cnt == 1 && err == nil {
		date.FromTime(time.Now().AddDate(delta, 0, 0))
		return
	}

	return date, false
}
