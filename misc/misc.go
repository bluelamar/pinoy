package misc

import (
	"fmt"
	"log"
	"strconv"
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

func XtractIntField(fieldName string, vmap *map[string]interface{}) int {
	num := int(0)
	if istr, ok := (*vmap)[fieldName].(string); ok {
		if inum, err := strconv.Atoi(istr); err == nil {
			return inum
		}
	}
	if ival, ok := (*vmap)[fieldName].(int); ok {
		num = ival
	} else if ival, ok := (*vmap)[fieldName].(int32); ok {
		num = int(ival)
	} else if fval, ok := (*vmap)[fieldName].(float64); ok {
		num = int(fval)
	}
	return num
}

func XtractFloatField(fieldName string, vmap *map[string]interface{}) float64 {
	num := float64(0)
	if fval, ok := (*vmap)[fieldName].(float64); ok {
		num = fval
	} else if ival, ok := (*vmap)[fieldName].(int); ok {
		num = float64(ival)
	}
	return num
}
