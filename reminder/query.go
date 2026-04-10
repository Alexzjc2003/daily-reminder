package reminder

import (
	"fmt"
	"time"
)

type QueryParam struct {
	From   *time.Time
	To     *time.Time
	Traits []string
	Mode   QueryMode
}

type QueryMode string

const (
	ModeRegular QueryMode = "regular"
	ModeExpand  QueryMode = "expand"
)

type QueryRecord struct {
	Record *ReminderRecord
	Date   ReminderDate
	Time   time.Time
}

type QueryResult QueryRecord

type QueryPredicate func(QueryRecord) bool
type QueryExpansion func(ReminderRecord) []QueryRecord

type ReminderQuery struct {
	Predicates []QueryPredicate
	Expansion  QueryExpansion
}

func (q *ReminderQuery) Build(param QueryParam) {
	if param.From != nil {
		q.From(*param.From)
	}

	if param.To != nil {
		q.To(*param.To)
	}

	if len(param.Traits) > 0 {
		q.Predicates = append(q.Predicates, func(d QueryRecord) bool {
			return hasAllTraits(d.Record.Traits, param.Traits)
		})
	}

	switch param.Mode {
	case ModeExpand:
		q.Expand()
	case ModeRegular:
		q.Exact()
	}
}

func (q *ReminderQuery) From(tm time.Time) *ReminderQuery {
	q.Predicates = append(q.Predicates, func(d QueryRecord) bool {
		if d.Time.IsZero() {
			return false
		}
		return !d.Time.Before(dayStart(tm))
	})
	return q
}

func (q *ReminderQuery) To(tm time.Time) *ReminderQuery {
	q.Predicates = append(q.Predicates, func(d QueryRecord) bool {
		if d.Time.IsZero() {
			return false
		}
		return !d.Time.After(dayStart(tm))
	})
	return q
}

// Date expansion is rather complicated. But a simple rule is that:
// yyyy = 0000: expand to -5 ~ +5 years based on this year
//   mm = 00  : expand to each month
//   dd = 00  : expand to each day

func (q *ReminderQuery) Expand() *ReminderQuery {
	q.Expansion = func(r ReminderRecord) (rs []QueryRecord) {
		switch r.Date.Kind() {
		case FULLDATE:
			rs = append(rs, makeQueryRecord(&r, r.Date))
		case NODATE:
			rs = append(rs, makeQueryRecord(&r, r.Date))
		case DAYONLY, MONTHONLY, MONTHDAY, YEARONLY, YEARDAY, YEARMONTH:
			rs = append(rs, expandPartialDate(r)...)
		default:
			panic(fmt.Sprintf("unexpected reminder.DateKind: %#v", r.Date.Kind()))
		}
		return
	}
	return q
}

func (q *ReminderQuery) Exact() *ReminderQuery {
	q.Expansion = func(r ReminderRecord) (rs []QueryRecord) {
		switch r.Date.Kind() {
		case FULLDATE:
			rs = append(rs, makeQueryRecord(&r, r.Date))
		}
		return
	}
	return q
}

func (q ReminderQuery) Apply(r Reminder) (results []QueryResult) {
	var candidates []QueryRecord

	// 1. expand
	for _, d := range r.Dates {
		candidates = append(candidates, q.Expansion(d)...)
	}

	// 2. filter
	predicate := func(r QueryRecord) bool {
		for _, p := range q.Predicates {
			if !p(r) {
				return false
			}
		}
		return true
	}

	for _, c := range candidates {
		if predicate(c) {
			results = append(results, QueryResult(c))
		}
	}

	return
}

func dayStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func hasAllTraits(recordTraits []string, required []string) bool {
	if len(required) == 0 {
		return true
	}
	recordSet := map[string]struct{}{}
	for _, t := range recordTraits {
		if t == "" {
			continue
		}
		recordSet[t] = struct{}{}
	}
	for _, req := range required {
		if req == "" {
			continue
		}
		if _, ok := recordSet[req]; !ok {
			return false
		}
	}
	return true
}

func expandPartialDate(r ReminderRecord) []QueryRecord {
	baseYears := yearsForExpansion(r.Date.Year)
	months := monthsForExpansion(r.Date.Month)

	var out []QueryRecord
	for _, y := range baseYears {
		for _, m := range months {
			days := daysForExpansion(y, m, r.Date.Day)
			for _, d := range days {
				expanded := ReminderDate{Year: uint(y), Month: uint(m), Day: uint(d)}
				out = append(out, makeQueryRecord(&r, expanded))
			}
		}
	}
	return out
}

func makeQueryRecord(record *ReminderRecord, date ReminderDate) QueryRecord {
	var t time.Time
	if date.Kind() == FULLDATE && isValidDate(int(date.Year), time.Month(date.Month), int(date.Day)) {
		t = time.Date(int(date.Year), time.Month(date.Month), int(date.Day), 0, 0, 0, 0, time.Local)
	}
	return QueryRecord{
		Record: record,
		Date:   date,
		Time:   t,
	}
}

func yearsForExpansion(year uint) []int {
	if year != 0 {
		return []int{int(year)}
	}
	nowYear := time.Now().Year()
	var years []int
	for y := nowYear - 5; y <= nowYear+5; y++ {
		years = append(years, y)
	}
	return years
}

func monthsForExpansion(month uint) []time.Month {
	if month != 0 {
		return []time.Month{time.Month(month)}
	}
	var months []time.Month
	for m := time.January; m <= time.December; m++ {
		months = append(months, m)
	}
	return months
}

func daysForExpansion(year int, month time.Month, day uint) []int {
	if day != 0 {
		if isValidDate(year, month, int(day)) {
			return []int{int(day)}
		}
		return nil
	}
	limit := daysInMonth(year, month)
	var days []int
	for d := 1; d <= limit; d++ {
		days = append(days, d)
	}
	return days
}

func daysInMonth(year int, month time.Month) int {
	t := time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local)
	return t.Day()
}

func isValidDate(year int, month time.Month, day int) bool {
	if day <= 0 {
		return false
	}
	return day <= daysInMonth(year, month)
}
