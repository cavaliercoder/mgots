package mgots

import (
	"time"
)

type DataPoint struct {
	Timestamp time.Time
	Value     interface{}
}

type DataPoints []DataPoint

type dataPage struct {
	PageId     interface{} `bson:"_id"` // ID of the time series page
	SeriesId   interface{} // ID of the series that this page belows to
	StartTime  time.Time   // Timestamp of the last entry in the previous page
	EndTime    time.Time   // Timestamp of the first entry in the next page
	Timestamps []time.Time // Array of timestamps for all entries in the page
	Values     interface{} // time series values for this page
}

type seriesCursor struct {
	SeriesId      interface{} `bson:"_id"` // ID of the series described by this cursor
	LastPage      interface{} // StartTime (ID) of the page pointed to by this cursor
	NextSlotId    int         // Index of the next available slot in the page pointed to by this cursor
	LastValue     interface{} // Value of the last entry in the series described by this cursor
	LastValueTime time.Time   // Timestamp of the last entry in the series described by this cursor
}

var cursorSuffix = "_cursors"
