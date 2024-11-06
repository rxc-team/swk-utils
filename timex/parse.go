package timex

import (
	"fmt"
	"time"
)

var (
	timeFormats = []string{
		"20060102",   // yyyyMMdd
		"2006-01-02", // yyyy-mm-dd
		"2006-1-02",  // yyyy-m-dd
		"2006-01-2",  // yyyy-mm-d
		"2006-1-2",   // yyyy-m-d
		"2006/01/02", // yyyy/mm/dd
		"2006/1/02",  // yyyy/m/dd
		"2006/01/2",  // yyyy/mm/d
		"2006/1/2",   // yyyy/m/d
		"2006.01.02", // yyyy.mm.dd
		"2006.1.02",  // yyyy.m.dd
		"2006.01.2",  // yyyy.mm.d
		"2006.1.2",   // yyyy.m.d
	}
)

func ToTimeE(s string) (d time.Time, e error) {
	return parseDateWith(s, time.UTC, timeFormats)
}

func ToTime(s string) time.Time {
	t, _ := parseDateWith(s, time.UTC, timeFormats)
	return t
}

func parseDateWith(s string, location *time.Location, formats []string) (d time.Time, e error) {

	for _, format := range formats {
		if d, e = time.Parse(format, s); e == nil {

			// Some time formats have a zone name, but no offset, so it gets
			// put in that zone name (not the default one passed in to us), but
			// without that zone's offset. So set the location manually.
			if location == nil {
				location = time.Local
			}
			year, month, day := d.Date()
			hour, min, sec := d.Clock()
			d = time.Date(year, month, day, hour, min, sec, d.Nanosecond(), location)

			return
		}
	}
	return d, fmt.Errorf("unable to parse date: %s", s)
}
