# mgots - Mango TS [![Build Status](https://travis-ci.org/cavaliercoder/mgots.svg)](https://travis-ci.org/cavaliercoder/mgots)

*A Time Series data model API for Go and MongoDB*

Mango TS is a [Google Go](https://golang.org/) package built on the
[mgo package](https://labix.org/mgo) which implements an optimized model for
periodic and nonperiodic Time Series data stored in [MongoDB](https://www.mongodb.org/).

## Time series data

From [Wikipedia](http://en.wikipedia.org/wiki/Time_series):

> A time series is a sequence of data points, typically consisting of 
  successive measurements made over a time interval.

### Periodic data

Period data appears in cronological order, at regular, predefined intervals.

Examples:

 * Performance monitoring metrics every minute
 * Scheduled tasks every hour

Periodic data is arrange into pages, each representing a linear segment of
time, with a _slot_ preallocated for each expected data point. For example, if
the data interval is one minute and the page size is one hour, each page will
contain 60 preallocated slots; one for each minute of the hour.

The model implemented in mgots is loosely based on [Sandeep Parikh's](http://blog.mongodb.org/post/65517193370/schema-design-for-time-series-data-in-mongodb)
time series schema design.

### Nonperiodic data

Nonperiodic data appears in cronological order, but at irregular intervals.
The data model used for periodic data is not appropriate for datasets of this
kind as a significant amount of storage is wasted on preallocated intervals for
which no data is entered.

Examples:

 * Monitoring events at random
 * User transactions at random

Nonperiodic data is arranged into a page until a configurable page size is
exhausted (E.g. 4096 bytes) when a new page is allocated. Each page represents
a linear segment of time, starting from the last entry point in the previous
page, until the first entry point of the next page.

The model implemented in mgots is loosely based on [Sylvain Wallez's](http://bluxte.net/musings/2015/01/21/efficient-storage-non-periodic-time-series-mongodb)
model used in [Actoboard](http://www.actoboard.com/).

## Page sizes
A number of facters contribute to optimizing page sizes. These include but are
no limited to:

 * Disk alignment
 * Alignment in memory of memory-mapped storage
 * MongoDB's [Power of 2](http://docs.mongodb.org/manual/core/storage/#power-of-2-sized-allocations)
   allocation strategy
 * Memory consumption and network utilization when querying pages and returning data

