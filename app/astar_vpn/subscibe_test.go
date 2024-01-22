package astar_vpn

import (
	"testing"
)

func TestUpdate(m *testing.T) {
	writeFile("app/astar_vpn/subscribe", Update())
}
