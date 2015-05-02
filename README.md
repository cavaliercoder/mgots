# mgots - Mango TS

*A Time Series data model API for Go and MongoDB*

Mango TS is a [Google Go](https://golang.org/) package built on the
[mgo package](https://labix.org/mgo) which implements an optimized model for
periodic and nonperiodtic Time Series data stored in [MongoDB](https://www.mongodb.org/).

## Time series data

From [Wikipedia](http://en.wikipedia.org/wiki/Time_series):

> A time series is a sequence of data points, typically consisting of 
  successive measurements made over a time interval.

## Periodic data

Period data appears in cronological order, at regular, predefined intervals.

Examples:

 * Performance metric monitors
 * Scheduled tasks

## Nonperiodic data

Nonperiodic data appears in cronological order, but at irregular intervals.
The data model used for periodic data is not appropriate for datasets of this
kind as a significant amount of storage is wasted on intervals for which no
data is entered.

Examples:

 * Monitoring events
 * User transactions

The model implemented in mgots is loosely based on [Sylvain Wallez's](http://bluxte.net/musings/2015/01/21/efficient-storage-non-periodic-time-series-mongodb)
model used in [Actoboard](http://www.actoboard.com/).
