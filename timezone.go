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

type TZInfo struct {
	Latitude  float64
	Longitude float64
	Location  string
	ZoneName  string
	IsDST     bool
	Offset    int
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
func (a *App) GetTimezoneInfo(lat, lon float64) (*TZInfo, error) {
	res, err := a.tzc.Search(lat, lon)
	if err != nil {
		return nil, err
	}
	a.logger.Debug(fmt.Sprintf("Found time zone info for coordinates lat=%f lon=%f", lat, lon))

	tzi := TZInfo{
		Latitude:  lat,
		Longitude: lon,
		Location:  res.Name,
	}

	loc, err := time.LoadLocation(tzi.Location)
	if err != nil {
		return nil, err
	}
	// Declaring t for Zone method
	t := time.Now().In(loc)

	// Calling Zone() method
	tzi.ZoneName, tzi.Offset = t.Zone()
	tzi.IsDST = t.IsDST()
	a.logger.Debug(fmt.Sprintf("Found time zone info for coordinates lat=%f lon=%f", lat, lon), "time zone info", tzi)

	return &tzi, nil
}

// GetCurrentTimeOffset returns the time offset in seconds for the given coordinates
// or zero if no time zone info may be obtained from coordinates.
func (a *App) GetCurrentTimeOffset(lat, lon float64) int {
	tzi, err := a.GetTimezoneInfo(lat, lon)
	if err != nil {
		a.logger.Error(err.Error())
		return 0
	}
	return tzi.Offset
}
