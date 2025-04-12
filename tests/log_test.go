package tests

import (
	"errors"
	"testing"

	"github.com/HamsterTunnel/core/log"
	"github.com/fatih/color"
)

func TestPrintAllLogs(t *testing.T) {
	color.NoColor = false

	t.Log("Esecuzione log demo:\n")

	log.Message("Questo è un messaggio informativo")
	log.Warning("Questo è un warning")
	log.Error("Questo è un errore", errors.New("qualcosa è andato storto"))
	log.Success("Operazione completata con successo")
}
