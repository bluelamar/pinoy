package misc

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bluelamar/pinoy/config"
	"github.com/bluelamar/pinoy/database"
)

var timeLocale *time.Location
var numReg *regexp.Regexp

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
		istr = strings.TrimSpace(istr)
		if inum, err := strconv.Atoi(istr); err == nil {
			return inum
		} else {
			log.Println("xtract-int:WARN: Parse err=", err, " : for string=", istr)
			return num
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
	if fstr, ok := (*vmap)[fieldName].(string); ok {
		fstr = strings.TrimSpace(fstr)
		if f, err := strconv.ParseFloat(fstr, 64); err == nil {
			return f
		} else {
			log.Println("xtract-float:WARN: Parse err=", err, " : field=", fieldName, " : value=", fstr)
			return num
		}
	}
	if fval, ok := (*vmap)[fieldName].(float64); ok {
		num = fval
	} else if ival, ok := (*vmap)[fieldName].(int); ok {
		num = float64(ival)
	}
	return num
}

// FilterUsageMapByField returns nil for entry that does not contain field
func FilterUsageMapByField(uMap map[string]interface{}, field string) *map[string]interface{} {
	id := ""
	name, exists := uMap[field]
	if !exists {
		return nil
	}
	id = name.(string)
	if id == "" {
		// ignore this record
		return nil
	}

	return &uMap
}

// CleanupDbUsage entries from dbName, only entries containing the field
func CleanupDbUsage(dbName, field string) error {
	// remove all entities from specified db
	resArray, err := database.DbwReadAll(dbName)
	if err != nil {
		log.Println(`misc.CleanupUsage:ERROR: db readall: err=`, err)
		return err
	}

	for _, v := range resArray {
		vm := v.(map[string]interface{})
		ru := FilterUsageMapByField(vm, field)
		if ru == nil {
			continue
		}
		if err := database.DbwDelete(dbName, ru); err != nil {
			log.Println(`misc.CleanupUsage:ERROR: db delete: err=`, err, ` : usage=`, ru)
		}
	}
	return nil
}

// CopyDbUsage entries from fromDB to the toDB, but only entries containing the field
func CopyDbUsage(fromDB, toDB, field string) error {
	// copy each entity from fromDB to the toDB
	resArray, err := database.DbwReadAll(fromDB)
	if err != nil {
		log.Println(`misc.CopyDbUsage:ERROR: db readall: err=`, err)
		return err
	}
	for _, v := range resArray {
		vm := v.(map[string]interface{})
		ru := FilterUsageMapByField(vm, field)
		if ru == nil {
			continue
		}
		err = database.DbwUpdate(toDB, (*ru)[field].(string), ru)
		if err != nil {
			log.Println("misc.CopyDbUsage:ERROR: Failed to update db=", toDB, " : for usage=", ru, " : err=", err)
			return err
		}
	}

	return nil
}

// StripMonPrefix will strip configured monetary prefix, ex: "$""
func StripMonPrefix(str string) string {
	mp := config.GetConfig().MonetarySymbol
	str = strings.ReplaceAll(str, mp, "")

	if numReg == nil {
		reg, err := regexp.Compile("[^0-9.]+")
		if err != nil {
			log.Println("misc.StripMonPrefix:ERROR: Failed to setup regex: err=", err)
		} else {
			numReg = reg
		}
	}
	if numReg != nil {
		str = numReg.ReplaceAllString(str, "")
	}
	return strings.TrimSpace(str)
}
