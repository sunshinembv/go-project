package models

import (
	"encoding/json"
	"fmt"
)

type Status int

const (
	NewStatus = iota
	InProgressStatus
	CompletedStatus
)

var statuses = []string{"new", "inProgress", "completed"}

func (s Status) String() string {
	return statuses[s]
}

func ParseStatus(status string) (Status, error) {
	switch status {
	case "new":
		return NewStatus, nil
	case "inProgress":
		return InProgressStatus, nil
	case "completed":
		return CompletedStatus, nil
	default:
		return -1, fmt.Errorf("unknown status: %s", status)
	}
}

func (s *Status) Scan(value any) error {
	switch v := value.(type) {
	case string:
		parsed, err := ParseStatus(v)
		if err != nil {
			return err
		}
		*s = parsed
	case []byte:
		parsed, err := ParseStatus(string(v))
		if err != nil {
			return err
		}
		*s = parsed
	default:
		return fmt.Errorf("cannot scan %T into Status", value)
	}
	return nil
}

func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *Status) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	parsed, err := ParseStatus(str)
	if err != nil {
		return err
	}

	*s = parsed
	return nil
}
