package datatype

import "testing"

func TestNetworkConfigScanString(t *testing.T) {
	var cfg NetworkConfig
	err := cfg.Scan(`{"networkStates":[{"id":"wan","name":"公网","type":"wan","host":""}],"activeNetworkStateId":"wan"}`)
	if err != nil {
		t.Fatalf("string scan failed: %v", err)
	}
	if len(cfg.NetworkStates) == 0 {
		t.Fatal("empty states")
	}
}

func TestNetworkConfigScanBytes(t *testing.T) {
	var cfg NetworkConfig
	err := cfg.Scan([]byte(`{"networkStates":[{"id":"wan","name":"公网","type":"wan","host":""}],"activeNetworkStateId":"wan"}`))
	if err != nil {
		t.Fatalf("bytes scan failed: %v", err)
	}
}
