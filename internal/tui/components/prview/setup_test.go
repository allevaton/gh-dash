package prview

import (
	"os"
	"testing"

	zone "github.com/lrstanley/bubblezone/v2"
)

func TestMain(m *testing.M) {
	zone.NewGlobal()
	zone.SetEnabled(false)
	os.Exit(m.Run())
}
