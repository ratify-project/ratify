package cmd

import "testing"

func TestDiscover(_ *testing.T) {
	// validate discover command does not crash
	_ = discover((discoverCmdOptions{
		subject:       "localhost:5000/net-monitor:v1",
		artifactTypes: []string{""},
	}))
}
