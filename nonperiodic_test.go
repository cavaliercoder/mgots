package mgots

import (
	"gopkg.in/mgo.v2/bson"
	"testing"
	"time"
)

type testData struct {
	Sequence int
	Padding  string
}

type timeRange struct {
	MinTime         time.Time
	MaxTime         time.Time
	ExpectedResults int
}

func TestNPOldData(t *testing.T) {
	database := DBConnect()
	name := "test_np_old_data"

	// Create a nonperiodic collection
	collection, err := NewNonperiodicCollection(database, name, testPageSize)
	if err != nil {
		t.Errorf(err.Error())
	}

	// Create a new series in the collection
	seriesId := bson.NewObjectId()
	startTime := time.Now().AddDate(-1, 0, 0)
	err = collection.CreateSeries(seriesId, startTime)
	if err != nil {
		t.Errorf(err.Error())
	}

	// Create a reasonable series entry
	value := "x"
	timestamp := time.Now()
	err = collection.Append(seriesId, timestamp, value)
	if err != nil {
		t.Errorf(err.Error())
	}

	// Make sure older entries always fail
	for i := 1; i < 10; i++ {
		timestamp = timestamp.Add(time.Duration(0-i) * time.Microsecond)
		err = collection.Append(seriesId, timestamp, value)
		if err != ErrTooOld {
			t.Errorf("Timestamp was too old but did not cause an error")
		}
	}
}

func TestNPAppending(t *testing.T) {
	database := DBConnect()
	name := "test_np_appending"

	// Create a nonperiodic collection
	collection, err := NewNonperiodicCollection(database, name, testPageSize)
	if err != nil {
		t.Errorf(err.Error())
	}

	// Create a new series in the collection
	seriesId := bson.NewObjectId()
	startTime := time.Now().AddDate(-1, 0, 0)
	err = collection.CreateSeries(seriesId, startTime)
	if err != nil {
		t.Errorf(err.Error())
	}

	// Add sequencial data to the series
	entryCount := 1000
	for i := 0; i < entryCount; i++ {
		timestamp := startTime.Add(time.Duration(i*60) * time.Minute)
		value := testData{i, "A little bit of padding."}

		err := collection.Append(seriesId, timestamp, value)
		if err != nil {
			t.Errorf(err.Error())
		}
	}

	// Fetch all data
	minTime := time.Now().AddDate(-2, 0, 0)
	maxTime := time.Now().AddDate(2, 0, 0)
	data, err := collection.Range(seriesId, minTime, maxTime)
	dataLen := len(data)
	if dataLen != entryCount {
		t.Errorf("Expected %d data entries to be returned. Got %d.", entryCount, dataLen)
	}

	// Ensure data is in order
	var lastEntry DataPoint = nil
	for i := 0; i < dataLen; i++ {
		entry := data[i]

		// Validate sequence id
		var rData testData
		err = entry.GetValue(&rData)
		if err != nil {
			t.Errorf("Error getting range data for %d: %s", i, err.Error())
		} else {
			if i != rData.Sequence {
				t.Errorf("Expected data sequence %d. Got %d.", i, rData.Sequence)
			}
		}

		// validate date order
		if lastEntry != nil && lastEntry.Timestamp().After(entry.Timestamp()) {
			t.Errorf("Sequence %d is more recent than sequence %d.", i-1, i)
		}

		lastEntry = entry
	}
}

func TestNPPaging(t *testing.T) {
	database := DBConnect()
	name := "test_np_paging"

	// Create a nonperiodic collection
	collection, err := NewNonperiodicCollection(database, name, testPageSize)
	if err != nil {
		t.Errorf(err.Error())
	}

	// Create a new series in the collection
	seriesId := bson.NewObjectId()
	startTime := time.Now().AddDate(-1, 0, 0)
	err = collection.CreateSeries(seriesId, startTime)
	if err != nil {
		t.Errorf(err.Error())
	}

	// Add sequencial data to the series
	// One for every half day
	endTime := startTime.AddDate(8, 0, 0)
	for timestamp := startTime; timestamp.Before(endTime); timestamp = timestamp.AddDate(0, 0, 1) {
		value := timestamp // just use the timestamp as an arbitrary value
		err = collection.Append(seriesId, timestamp, value)
		if err != nil {
			t.Errorf(err.Error())
		}
	}

	// Fetch the created pages
	var pages []dataPage
	err = database.C(name).Find(bson.M{}).Sort("starttime").All(&pages)
	if err != nil {
		t.Fatalf("Failed to fetch pages")
	}

	// validate start and end dates of each pages
	var lastPage *dataPage = nil
	pageCount := len(pages)
	for i, page := range pages {
		slots := len(page.Timestamps)
		if slots == 0 {
			t.Errorf("Page %d has no slots", i)
		} else {

			// test page StartTimes
			if i == 0 {
				// for the first page, page.StartTime should equal the date of the oldest slot
				if page.StartTime != page.Timestamps[slots-1] {
					t.Errorf("StartTime of the first page is not equal to the first timestamp in the page.")
				}
			} else {
				// start time of this page should be equal to the last timestamp in the previous page
				if page.StartTime != lastPage.Timestamps[0] {
					t.Errorf("StartTime of page %d (%s) is not equal to the last timestamp of page %d (%s)", i, page.StartTime.Format(layout), i-1, lastPage.Timestamps[0].Format(layout))
					t.Errorf("Last page start: %s\n\n", lastPage.StartTime.Format(layout))
				}
			}

			// test page EndTimes
			if i == pageCount-1 {
				// get latest slot in this page
				var zeroTime time.Time
				i := 0
				for ; page.Timestamps[i] == zeroTime; i++ {
				}

				// page.EndTime should equal the latest slot
				latestSlot := page.Timestamps[i]
				if page.EndTime != latestSlot {
					t.Errorf("EndTime of the last page %d (%s) is not equal to the last timestamp in the page (%s).", i, page.EndTime.Format(layout), latestSlot.Format(layout))
				}
			} else {
				// get earliest slot in the next page
				nextPage := pages[i+1]
				nextPageSlotCount := len(nextPage.Timestamps)
				earliestSlot := nextPage.Timestamps[nextPageSlotCount-1]

				// page.EndTime should be equal to the earliest slot in the next page
				if page.EndTime != earliestSlot {
					t.Errorf("EndTime of page %d (%s) is not equal to the earliest timestamp in the next page (%s).", i, page.EndTime.Format(layout), earliestSlot.Format(layout))
				}
			}

			lastPage = &pages[i]
		}
	}
}

func TestNPLatest(t *testing.T) {
	database := DBConnect()
	name := "test_np_latest"

	// Create a nonperiodic collection
	collection, err := NewNonperiodicCollection(database, name, testPageSize)
	if err != nil {
		t.Errorf(err.Error())
	}

	// Create a new series in the collection
	seriesId := bson.NewObjectId()
	startTime := time.Now().AddDate(-1, 0, 0)
	err = collection.CreateSeries(seriesId, startTime)
	if err != nil {
		t.Errorf(err.Error())
	}

	// Add sequencial data to the series and validate the current value
	entryCount := 1000
	for i := 0; i < entryCount; i++ {
		// next data point
		// truncate the timestamp to milliseconds to account for loss of
		// precision in MongoDB
		timestamp := startTime.Add(time.Duration(i*60) * time.Minute).Truncate(time.Millisecond)
		value := testData{i, "A little bit of padding."}

		// append to time series
		err := collection.Append(seriesId, timestamp, value)
		if err != nil {
			t.Errorf(err.Error())
		}

		// validate latest value function
		latest, err := collection.Latest(seriesId)
		if err != nil {
			t.Errorf(err.Error())
		} else {
			if !latest.Timestamp().Equal(timestamp) {
				t.Errorf("Latest timestamp (%s) does not match the most recently appended timestamp (%s)", latest.Timestamp().Format(layout), timestamp.Format(layout))
			}

			var rData testData
			err = latest.GetValue(&rData)
			if err != nil {
				t.Errorf("Error unmarshalling range data: %s", err.Error())
			} else {
				if rData.Sequence != i {
					t.Errorf("Latest value sequence %d does not match actual sequence %d", rData.Sequence, i)
				}
			}
		}

		// validate update function
		value.Padding = "Updated value"
		err = collection.Update(seriesId, value)
		if err != nil {
			t.Errorf("Error updating most recent value: %s", err.Error())
		}

		var uValue testData
		latest, err = collection.Latest(seriesId)
		if err != nil {
			t.Errorf("Error getting updated value: %s", err.Error())
		} else {
			if err = latest.GetValue(&uValue); err != nil {
				t.Errorf("Error reading updated value: %s", err.Error())
			} else {
				if uValue.Padding != value.Padding {
					t.Errorf("Value does not appear to have been updated by Update()")
				}
			}
		}
	}
}

func TestNPRanging(t *testing.T) {
	database := DBConnect()
	name := "test_np_ranging"

	// Create a nonperiodic collection
	collection, err := NewNonperiodicCollection(database, name, testPageSize)
	if err != nil {
		t.Errorf(err.Error())
	}

	// Create a new series in the collection
	seriesId := bson.NewObjectId()
	startTime := time.Now().AddDate(-1, 0, 0)
	err = collection.CreateSeries(seriesId, startTime)
	if err != nil {
		t.Errorf(err.Error())
	}

	// Add sequencial data to the series
	// 1440 entry points, one for every minute of the day
	endTime := startTime.AddDate(0, 0, 1)
	for timestamp := startTime; timestamp.Before(endTime); timestamp = timestamp.Add(time.Duration(1) * time.Minute) {
		value := timestamp // just use the timestamp as an arbitrary value
		err = collection.Append(seriesId, timestamp, value)
		if err != nil {
			t.Errorf(err.Error())
		}
	}

	// Generate some test ranges
	ranges := []timeRange{
		{startTime, endTime, 1439},
		{startTime.AddDate(0, 0, -1), endTime.AddDate(0, 0, 1), 1440},
		{startTime, startTime.Add(time.Duration(120) * time.Minute), 120},
		{startTime.Add(time.Duration(50) * time.Minute), startTime.Add(time.Duration(150) * time.Minute), 100},
	}

	// Test each range
	for i, queryRange := range ranges {
		// fetch data for this time range
		data, err := collection.Range(seriesId, queryRange.MinTime, queryRange.MaxTime)
		if err != nil {
			t.Errorf("Error for range %#v:\n%s", queryRange, err.Error())
		}

		// Validate result count
		dataLen := len(data)
		if dataLen != queryRange.ExpectedResults {
			t.Errorf("Expected %d results, got %d for range %d.\n", queryRange.ExpectedResults, dataLen, i)
		}

		// Validate range entry dates
		for x, entry := range data {
			if entry.Timestamp().Before(queryRange.MinTime) {
				t.Errorf("Result %d (%s) is earlier than the range minimum (%s)", x, entry.Timestamp().Format(layout), queryRange.MinTime.Format(layout))
			}

			if entry.Timestamp().After(queryRange.MaxTime) {
				t.Errorf("Result %d (%s) is later than the range maximum (%s)", x, entry.Timestamp().Format(layout), queryRange.MaxTime.Format(layout))
			}
		}
	}
}
