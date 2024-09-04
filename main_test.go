package tbb_test

import (
	"fmt"
	"github.com/dusted-go/logging/prettylog"
	"log/slog"
	"os"
	"testing"
)

var logger *slog.Logger

func TestMain(m *testing.M) {
	setup()
	teardown(m.Run())
}

func setup() {
	logger = slog.New(prettylog.NewHandler(&slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelInfo,
	})).With("testing", "on")
	logger.Info("Setup tests...")
}

func teardown(code int) {
	logger.Info("Teardown tests...")

	if err := os.Remove("test/data/test.app.db"); err != nil {
		logger.Error(err.Error())
	}

	logger.Info(fmt.Sprintf("Exit code: %d", code), "code", code)
	os.Exit(code)
}
