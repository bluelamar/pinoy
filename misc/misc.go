package misc

import (
	"fmt"
	"log"
	"time"
)

var timeLocale *time.Location

func GetLocale() *time.Location {
	if timeLocale == nil {
		secondsEastOfUTC := int((8 * time.Hour).Seconds())

		timeLocale = time.FixedZone("Maynila Time", secondsEastOfUTC)
	}
	return timeLocale
}

func TimeNow() (string, time.Time) {
	locale := GetLocale()
	return TimeNowLocale(locale)
}

func TimeNowLocale(locale *time.Location) (string, time.Time) {
	var now time.Time
	if locale == nil {
		locale = GetLocale()
	}
	now = time.Now().In(locale)
	nowStr := fmt.Sprintf("%d-%02d-%02d %02d:%02d",
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute())
	return nowStr, now
}

func InitTime(timeZoneName string, hoursEastOfUTC time.Duration) {
	loc, err := time.LoadLocation(timeZoneName) // "Singapore")
	if err != nil {
		log.Println("misc:WARN: Failed to load singapore time location: Use default locale: +0800 UTC-8: err=", err)
		//Locale = time.FixedZone("UTC-8", 8*60*60)
		// secondsEastOfUTC := int((8 * time.Hour).Seconds())
		secondsEastOfUTC := int((hoursEastOfUTC * time.Hour).Seconds())
		// Locale = time.FixedZone("Maynila Time", secondsEastOfUTC)
		timeLocale = time.FixedZone(timeZoneName, secondsEastOfUTC)
	} else {
		timeLocale = loc
	}
}
