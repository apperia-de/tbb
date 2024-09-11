package tbb

import (
	"embed"
	"fmt"
	timezone "github.com/evanoberholster/timezoneLookup/v2"
	"os"
	"time"
)

//go:embed assets/timezone.data
var efs embed.FS

type TimeZoneInfo struct {
	Latitude  float64 `json:"latitude,omitempty"`  // Latitude the user sends for determining the user's current Time zone
	Longitude float64 `json:"longitude,omitempty"` // Longitude the user sends for determining the user's current Time zone
	Location  string  `json:"location,omitempty"`  // Location of the user's timezone
	ZoneName  string  `json:"zoneName,omitempty"`  // Zone name of the user's timezone
	Offset    int     `json:"offset,omitempty"`    // Time zone offset in seconds
	IsDST     bool    `json:"isDST,omitempty"`     // Whether the offset is in daylight saving time or normal time
}

func loadTimezoneCache() *timezone.Timezonecache {
	var (
		f, tempF *os.File
		tzc      timezone.Timezonecache
		data     []byte
		err      error
	)

	tempF, err = os.CreateTemp("", "timezone.data")
	if err != nil {
		panic(err)
	}
	defer os.Remove(tempF.Name())

	data, err = efs.ReadFile("assets/timezone.data")
	if err != nil {
		panic(err)
	}

	_, err = tempF.Write(data)
	if err != nil {
		panic(err)
	}

	f, err = os.Open(tempF.Name())
	if err != nil {
		panic(err)
	}

	if err = tzc.Load(f); err != nil {
		panic(err)
	}

	return &tzc
}

// GetTimezoneInfo returns the time zone info for the given coordinates if available.
func (tb *TBot) GetTimezoneInfo(lat, lon float64) (*TimeZoneInfo, error) {
	res, err := tb.tzc.Search(lat, lon)
	if err != nil {
		return nil, err
	}
	tb.logger.Debug(fmt.Sprintf("Found time zone info for coordinates lat=%f lon=%f", lat, lon))

	tzi := TimeZoneInfo{
		Latitude:  lat,
		Longitude: lon,
		Location:  res.Name,
	}

	loc, err := time.LoadLocation(tzi.Location)
	if err != nil {
		return nil, err
	}
	// Declaring tb for Zone method
	t := time.Now().In(loc)

	// Calling Zone() method
	tzi.ZoneName, tzi.Offset = t.Zone()
	tzi.IsDST = t.IsDST()
	tb.logger.Debug(fmt.Sprintf("Found time zone info for coordinates lat=%f lon=%f", lat, lon), "time zone info", tzi)

	return &tzi, nil
}

// GetCurrentTimeOffset returns the time offset in seconds for the given coordinates
// or zero if no time zone info may be obtained from coordinates.
func (tb *TBot) GetCurrentTimeOffset(lat, lon float64) int {
	tzi, err := tb.GetTimezoneInfo(lat, lon)
	if err != nil {
		tb.logger.Error(err.Error())
		return 0
	}
	return tzi.Offset
}
