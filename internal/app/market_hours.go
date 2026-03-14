package app

import "time"

var easternTZ *time.Location

func init() {
	var err error
	easternTZ, err = time.LoadLocation("America/New_York")
	if err != nil {
		easternTZ = time.FixedZone("EST", -5*60*60)
	}
}

// isUSMarketOpen returns true during regular US trading hours (9:30–16:00 ET, weekdays).
func isUSMarketOpen() bool {
	return isUSMarketOpenAt(time.Now())
}

func isUSMarketOpenAt(t time.Time) bool {
	t = t.In(easternTZ)
	weekday := t.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return false
	}
	minuteOfDay := t.Hour()*60 + t.Minute()
	return minuteOfDay >= 570 && minuteOfDay < 960 // 9:30 to 16:00
}

// isExtendedHours returns true during pre-market (4:00–9:30) or post-market (16:00–20:00) on weekdays.
func isExtendedHours() bool {
	return isExtendedHoursAt(time.Now())
}

func isExtendedHoursAt(t time.Time) bool {
	t = t.In(easternTZ)
	weekday := t.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return false
	}
	minuteOfDay := t.Hour()*60 + t.Minute()
	return (minuteOfDay >= 240 && minuteOfDay < 570) || // pre-market 4:00-9:30
		(minuteOfDay >= 960 && minuteOfDay < 1200) // post-market 16:00-20:00
}
