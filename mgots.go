package mgots

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type dataPage struct {
	PageId     interface{} `bson:"_id"` // ID of the time series page
	SeriesId   interface{} // ID of the series that this page belows to
	StartTime  time.Time   // Timestamp of the last entry in the previous page
	EndTime    time.Time   // Timestamp of the first entry in the next page
	Timestamps []time.Time `bson:",omitempty"` // Array of timestamps for all entries in the page
	Values     []bson.Raw  `bson:",omitempty"` // time series values for this page
	Padding    []byte      `bson:",omitempty"` // padding data to set initial page size
}

const (
	PAGE_HEADER_SIZE = 110 // page header size in bytes (as per Object.bsonsize())
	TIMESTAMP_SIZE   = 66  // Timestamp data type size in bytes (including encapsulation)
)

type seriesCursor struct {
	SeriesId      interface{} `bson:"_id"` // ID of the series described by this cursor
	LastPage      interface{} // StartTime (ID) of the page pointed to by this cursor
	NextSlotId    int         // Index of the next available slot in the page pointed to by this cursor
	LastValue     bson.Raw    `bson:",omitempty"` // Value of the last entry in the series described by this cursor
	LastValueTime time.Time   `bson:",omitempty"` // Timestamp of the last entry in the series described by this cursor
}

var cursorSuffix = "_cursors"
var timeZero = time.Unix(0, 0)
var bsonZero = bson.Raw{Kind: 0x0A, Data: []byte{}}

func BSONSize(v interface{}) int {
	s := []interface{}{v}
	b, err := bson.Marshal(s)
	if err != nil {
		panic(err)
	}

	return len(b)
}
