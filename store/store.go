package store

import (
	"time"

	"github.com/miniscruff/dashy/configs"
	"github.com/tidwall/gjson"
)

const timeFormat = time.ANSIC

type Store interface {
	StringOrVar(value string) string
	GetNextRun(feed *configs.FeedConfig) (time.Time, error)
	SetNextRun(feed *configs.FeedConfig, nextRun time.Time) error
	GetValues() (map[string]map[string]interface{}, error)
	SetValues(feed *configs.FeedConfig, values map[string]gjson.Result) error
}
