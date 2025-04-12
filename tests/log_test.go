
import (
	"errors"
	"testing"

	"github.com/HamsterTunnel/core/log"
	"github.com/fatih/color"
)

func TestPrintAllLogs(t *testing.T) {
	color.NoColor = false

	t.Log("Running log demo:\n")

	log.Message("This is an informational message")
	log.Warning("This is a warning")
	log.Error("This is an error", errors.New("something went wrong"))
	log.Error("This is an error without an error")
	log.Success("Operation completed successfully")
}
