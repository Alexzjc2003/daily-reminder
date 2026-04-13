package reminder

import (
	"fmt"
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

func (d *ReminderDate) ToString() string {
	return fmt.Sprintf("%04d/%02d/%02d", d.Year, d.Month, d.Day)
}

func (d *ReminderDate) ToTime() time.Time {
	if d.Kind() == FULLDATE {
		return time.Date(int(d.Year), time.Month(d.Month), int(d.Day), 0, 0, 0, 0, time.Local)
	} else {
		return time.Time{}
	}
}
