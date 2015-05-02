package mgots

import (
	"gopkg.in/mgo.v2"
	"time"
)

type collection struct {
	Database             *mgo.Database
	Name                 string
	CursorCollectionName string
	DBCollection         *mgo.Collection
	DBCursorCollection   *mgo.Collection
}

type Collection interface {
	CreateSeries(seriesId interface{}, startTime time.Time) error
	Append(seriesId interface{}, timestamp time.Time, value interface{}) error
	Range(seriesId interface{}, minTime time.Time, maxTime time.Time) (DataPoints, error)
	Latest(seriesId interface{}) (*DataPoint, error)
}
