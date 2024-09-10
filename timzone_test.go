package tbb_test

import (
	"github.com/apperia-de/tbb"
	"github.com/evanoberholster/timezoneLookup/v2"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestApp_GetTimezoneInfo(t *testing.T) {
	t.Run("Get TimezoneInfo for valid coordinates", func(t *testing.T) {

		cfg := tbb.LoadConfig("test/data/test.config.yml")
		app := tbb.New(tbb.WithConfig(cfg))

		tzi, err := app.GetTimezoneInfo(51.340847907357755, 12.377381803667586)
		assert.NoError(t, err)
		assert.NotNil(t, tzi)
		assert.Equal(t, 51.340847907357755, tzi.Latitude)
		assert.Equal(t, 12.377381803667586, tzi.Longitude)
	})

	t.Run("Get TimezoneInfo for invalid coordinates", func(t *testing.T) {
		cfg := tbb.LoadConfig("test/data/test.config.yml")
		app := tbb.New(tbb.WithConfig(cfg))

		tzi, err := app.GetTimezoneInfo(-125.123, 0)
		assert.Nil(t, tzi)
		assert.ErrorIs(t, err, timezoneLookup.ErrCoordinatesNotValid)
	})
}

func TestApp_GetCurrentTimeOffset(t *testing.T) {
	cfg := tbb.LoadConfig("test/data/test.config.yml")
	app := tbb.New(tbb.WithConfig(cfg))

	lat, lon := 51.340847907357755, 12.377381803667586
	tzi, err := app.GetTimezoneInfo(lat, lon)
	assert.NoError(t, err)
	offset := app.GetCurrentTimeOffset(lat, lon)
	if tzi.IsDST {
		// Summer
		assert.Equal(t, 7200, offset)
	} else {
		// Winter
		assert.Equal(t, 3600, offset)
	}
}
