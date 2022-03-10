package store

import (
	"time"

	"github.com/tidwall/gjson"
)

const timeFormat = time.ANSIC

type Store interface {
	StringOrVar(value string) string
	GetNextRun(feedName string) (time.Time, error)
	SetNextRun(feedName string, nextRun time.Time) error
	GetValues() (map[string]map[string]interface{}, error)
	SetValues(feedName string, values map[string]gjson.Result) error
}
