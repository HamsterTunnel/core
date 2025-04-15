package log

import (
	"errors"
	"testing"

	"github.com/fatih/color"
)

func TestPrintAllLogs(t *testing.T) {
	color.NoColor = false

	t.Log("Running log demo:\n")

	Message("This is a message")
	Info("This is a info ")
	Warning("This is a warning")
	Error("This is an error", errors.New("something went wrong"))
	Error("This is an error without an error", nil)
	Success("Operation completed successfully")
}
