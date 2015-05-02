package mgots

import (
	"errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"reflect"
	"strconv"
	"time"
	"unsafe"
)

type NonperiodicCollection struct {
	collection
	PageSize int64
}

// Errors
var ErrSeriesNotFound = errors.New("Series not found with the specified collection")
var ErrInvalidPageSize = errors.New("Pages size must be a positive integer")
var ErrDuplicateSeries = errors.New("Time series already exists with the specified ID")
var ErrValueTooLarge = errors.New("The size value specified exceeds the maximum Time Series page size.")
var ErrTooOld = errors.New("The timestamp of the specified value is older than the most recent entry or the series does not exist.")

func NewNonperiodicCollection(database *mgo.Database, name string, pageSize int64) (Collection, error) {
	// Validate page size
	// TODO: Determine minimum page size (including header and timestamps array)
	if pageSize < 1 {
		return nil, ErrInvalidPageSize
	}

	// Build collection struct
	collection := NonperiodicCollection{
		collection: collection{
			Database:             database,
			Name:                 name,
			CursorCollectionName: name + cursorSuffix,
		},
		PageSize: pageSize,
	}

	// Attach to MongoDB collections
	collection.DBCollection = database.C(name)
	collection.DBCursorCollection = database.C(collection.CursorCollectionName)

	return &collection, nil
}

func (c *NonperiodicCollection) CreateSeries(seriesId interface{}, startTime time.Time) error {

	// Does the series exist?
	count, err := c.DBCursorCollection.FindId(seriesId).Count()
	if err != nil {
		return err
	}

	if count != 0 {
		return ErrDuplicateSeries
	}

	// Create new time series cursor
	err = c.DBCursorCollection.Insert(seriesCursor{
		SeriesId: seriesId,
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *NonperiodicCollection) Latest(seriesId interface{}) (*DataPoint, error) {
	// Fetch last value from the series cursor
	var cursor seriesCursor
	err := c.DBCursorCollection.FindId(seriesId).One(&cursor)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, ErrSeriesNotFound
		}

		return nil, err
	}

	return &DataPoint{
		Timestamp: cursor.LastValueTime,
		Value:     cursor.LastValue,
	}, nil
}

func (c *NonperiodicCollection) Range(seriesId interface{}, minTime time.Time, maxTime time.Time) (DataPoints, error) {
	// search for matching pages
	var pages []dataPage
	err := c.DBCollection.Find(bson.M{
		"seriesid":  seriesId,
		"starttime": bson.M{"$lte": maxTime},
		"endtime":   bson.M{"$gte": minTime},
	}).Sort("starttime").All(&pages)

	if err != nil {
		return nil, err
	}

	// create capacity for result set
	// estimate len(pages[0].Timestamps)
	pagesLen := len(pages)
	resultsLen := 1
	if pagesLen > 0 {
		slots := len(pages[0].Timestamps)
		resultsLen = (1 + pagesLen) * (1 + slots)
	}

	// convert to map
	j := 0
	results := make(DataPoints, resultsLen)
	for _, page := range pages {
		values := page.Values.([]interface{})

		for i := len(page.Timestamps) - 1; i >= 0; i-- {
			timestamp := page.Timestamps[i]

			if (timestamp.Equal(minTime) || timestamp.After(minTime)) && (timestamp.Equal(maxTime) || timestamp.Before(maxTime)) {
				results[j] = DataPoint{
					timestamp,
					values[i],
				}
				j++
			}
		}
	}

	// truncate and return
	return results[0:j], nil
}

/*
 * value must have a consistent size
 */
func (c *NonperiodicCollection) Append(seriesId interface{}, timestamp time.Time, value interface{}) error {
	// Query to find the series cursor
	query := c.DBCursorCollection.Find(bson.M{
		"_id":           seriesId,
		"lastvaluetime": bson.M{"$lt": timestamp},
	})

	// Compile the change to apply to the cursor
	change := mgo.Change{
		Update: bson.M{
			"$set": bson.M{
				"lastvalue":     value,
				"lastvaluetime": timestamp,
			},
			"$inc": bson.M{
				"nextslotid": -1,
			},
		},
		ReturnNew: true,
	}

	// Search and update the series cursor
	var cursor seriesCursor
	_, err := query.Apply(change, &cursor)
	if err != nil {
		if err == mgo.ErrNotFound {
			return ErrTooOld
		}
		return err
	}

	// Create a new page if the NextSlotId is < 0
	if cursor.NextSlotId < 0 {
		// Fetch and update last page
		var lastPage dataPage
		_, err := c.DBCollection.Find(bson.M{
			"_id": cursor.LastPage,
		}).Apply(
			mgo.Change{
				Update: bson.M{
					"$set": bson.M{
						"endtime": timestamp,
					},
				},
				ReturnNew: false,
			}, &lastPage)

		if err != nil {
			if err != mgo.ErrNotFound {
				return err
			}

			// No previous page
			lastPage.EndTime = timestamp
		}

		// Create new page
		newPage := dataPage{
			PageId:    bson.NewObjectId(),
			SeriesId:  seriesId,
			StartTime: lastPage.EndTime, // End time, prior to updating the previous page
			EndTime:   timestamp,
		}

		// Create empty values for new page
		// TODO: Compute slots/page using BSON instead of Go values
		slots := int(c.PageSize / int64(unsafe.Sizeof(value)))
		if slots < 1 {
			return ErrValueTooLarge
		}

		// Preallocate null data into page slots
		newPage.Timestamps = make([]time.Time, slots)

		valType := reflect.SliceOf(reflect.TypeOf(value))
		preAlloc := reflect.MakeSlice(valType, slots, slots)
		newPage.Values = preAlloc.Interface()

		// Insert new page
		err = c.DBCollection.Insert(newPage)
		if err != nil {
			return err
		}

		// Update cursor in database
		err = c.DBCursorCollection.UpdateId(cursor.SeriesId, bson.M{
			"$set": bson.M{
				"lastpage":   newPage.PageId,
				"nextslotid": slots, // -1 for the entry we already added
			},
		})
		if err != nil {
			return err
		}

		// update cursor for next operation
		cursor.LastPage = newPage.PageId
		cursor.NextSlotId = slots
	}

	// Update the next page and slot with this data
	slotIdString := strconv.FormatInt(int64(cursor.NextSlotId), 10)
	err = c.DBCollection.UpdateId(cursor.LastPage, bson.M{
		"$set": bson.M{
			"endtime":                    timestamp,
			"timestamps." + slotIdString: timestamp,
			"values." + slotIdString:     value,
		},
	})

	if err != nil {
		return err
	}

	return nil
}
