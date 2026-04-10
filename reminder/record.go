package reminder

import (
	"fmt"
	"strings"
)

type ReminderRecord struct {
	Name   string
	Date   ReminderDate
	Traits []string
}

// Format ReminderRecord into string
func (r ReminderRecord) ToString() string {
	return fmt.Sprintf("%s:%s;%s", r.Date.ToString(), r.Name, strings.Join(r.Traits, ","))
}

// Parse ReminderRecord from string
func (r *ReminderRecord) FromString(str string) error {
	var parts []string

	if parts = strings.SplitN(str, ":", 2); len(parts) < 2 {
		return fmt.Errorf("missing `:`")
	}

	date, err := ParseDate(parts[0])
	if err != nil {
		return fmt.Errorf("failed to parse date: %v", err)
	}

	if parts = strings.SplitN(parts[1], ";", 2); len(parts) < 2 {
		return fmt.Errorf("missing `;`")
	}
	name := parts[0]

	var traits []string
	for trait := range strings.SplitSeq(parts[1], ",") {
		traits = append(traits, strings.TrimSpace(trait))
	}

	*r = ReminderRecord{
		Name:   name,
		Date:   date,
		Traits: traits,
	}

	return nil
}
