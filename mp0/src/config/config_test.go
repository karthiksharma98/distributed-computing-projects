package config

import (
	"testing"
)

func TestConfig(t *testing.T) {
	adds, _ := IPAddress()

	for _, addr := range adds {
		t.Log(addr)
	}
	port, _ := Port()

	t.Log(port)
}
