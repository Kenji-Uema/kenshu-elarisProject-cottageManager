package common

import "time"

type DateRangeQuery struct {
	From time.Time `form:"from" time_format:"2006-01-02" binding:"required"`
	To   time.Time `form:"to"   time_format:"2006-01-02" binding:"required"`
}
