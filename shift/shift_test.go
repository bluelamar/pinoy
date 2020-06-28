package shift

import (
	"fmt"
	"testing"
	"time"

	"github.com/bluelamar/pinoy/misc"
)

func TestCalcMonthDayOfMonth(t *testing.T) {
	// calcMonthDayOfMonth(t time.Time) (string, int)
	tests := []struct {
		year, month, day, hour, minute int
	}{
		{2000, 1, 1, 12, 0},    // create time for jan 1, 2000
		{2000, 12, 31, 23, 59}, // create time for dec 31, 2000
		{2000, 7, 6, 8, 10},    // create time for july 6, 2000
	}

	var tstr string
	for _, v := range tests {
		tstr = fmt.Sprintf("%d-%02d-%02d %02d:%02d", v.year, v.month, v.day, v.hour, v.minute)
		if tTime, err := time.ParseInLocation("2006-01-02 15:04", tstr, misc.GetLocale()); err == nil {
			month, dom := calcMonthDayOfMonth(tTime)
			if month != monthList[v.month-1] {
				t.Errorf("expected month %v but got %v", monthList[v.month-1], month)
			}
			if dom != v.day {
				t.Errorf("expected day %v but got %v", v.day, dom)
			}
		}

	}
}
func TestCalcShift(t *testing.T) {
	shiftList = []ShiftItem{ // set the global shifts
		{1, 8, 15},
		{2, 15, 23},
		{3, 23, 8},
	}

	// calcShift(time.Time) (int, int, int, time.Time) => day-of-the-year, hour-of-the-day, shift-number
	// year 2000 is a leap year
	tests := []struct {
		year, month, day, hour, minute, shift int
	}{
		{2000, 1, 1, 12, 0, 1},    // create time for jan 1, 2000
		{2000, 12, 31, 23, 59, 3}, // create time for dec 31, 2000
		{2000, 7, 6, 8, 0, 1},     // create time for july 6, 2000
		{2000, 2, 1, 14, 59, 1},
		{2000, 2, 28, 14, 59, 1},
		{2000, 2, 28, 15, 0, 2},
		{2000, 2, 28, 22, 59, 2},
		{2000, 12, 31, 23, 0, 3},
		{2000, 1, 1, 7, 59, 3},
		{2000, 3, 1, 7, 59, 3},
		{2001, 3, 1, 7, 59, 3}, // not a leap year
	}

	var tstr string
	for _, v := range tests {
		tstr = fmt.Sprintf("%d-%02d-%02d %02d:%02d", v.year, v.month, v.day, v.hour, v.minute)
		if tTime, err := time.ParseInLocation("2006-01-02 15:04", tstr, misc.GetLocale()); err == nil {
			doy, hod, shift, _ := calcShift(tTime)
			if shift != v.shift {
				t.Errorf("expected shift %v but got %v: hour %v", v.shift, shift, v.hour)
			}
			if hod != v.hour {
				t.Errorf("expected shift %v but got %v", v.hour, hod)
			}
			if v.month == 1 {
				if v.day != doy {
					t.Errorf("expected day of year %v but got %v: month %v", v.day, doy, v.month)
				}
			} else if v.month == 2 {
				if (v.day + 31) != doy {
					t.Errorf("expected day of year %v but got %v: month %v", (v.day + 31), doy, v.month)
				}
			} else if v.month == 3 {
				if v.year == 2000 {
					if (v.day + 31 + 29) != doy {
						t.Errorf("expected day of leap year %v but got %v: month %v", (v.day + 31 + 29), doy, v.month)
					}
				} else if (v.day + 31 + 28) != doy {
					t.Errorf("expected day of year %v but got %v: month %v", (v.day + 31 + 28), doy, v.month)
				}
			}
		}
	}
}

func TestAdjustDayForXOverShift(t *testing.T) {
	// AdjustDayForXOverShift returns day-of-year adjusted for shift that crosses over midnight
	// AdjustDayForXOverShift(year, dayOfYear, hourOfDay, shiftNum int) int
	// 3rd shift is the one crosses over to day before
	tests := []struct {
		year, day, hour, shift, adjDay int
	}{
		{2021, 365, 23, 3, 1},   // crossover
		{2020, 365, 23, 3, 366}, // doesnt crossover - leap year
		{2020, 1, 0, 3, 1},      // doesnt crossover
		{2020, 1, 7, 3, 1},      // doesnt crossover
		{2020, 365, 1, 3, 365},  // doesnt crossover
		{2020, 32, 8, 1, 32},
		{2020, 32, 22, 2, 32},
	}

	for _, v := range tests {
		d := AdjustDayForXOverShift(v.year, v.day, v.hour, v.shift)
		if d != v.adjDay {
			t.Errorf("expected day of year %v but got %v: hour=%v year=%v", v.adjDay, d, v.hour, v.year)
		}
	}
}
