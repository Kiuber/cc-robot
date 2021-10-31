package ctime

import (
	"time"
)

func SecToStr(sec int64) string {
	unixTimeUTC := time.Unix(sec, 0)

	unitTimeInRFC3339 := unixTimeUTC.Format(time.RFC3339)
	return unitTimeInRFC3339
}
