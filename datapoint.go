package mgots

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type dataPoint struct {
	timestamp time.Time
	value     bson.Raw
}

type DataPoint interface {
	Timestamp() time.Time
	GetValue(v interface{}) error
}

type DataPoints []DataPoint

func (c *dataPoint) Timestamp() time.Time {
	return c.timestamp
}

func (c *dataPoint) GetValue(v interface{}) error {
	return c.value.Unmarshal(v)
}
