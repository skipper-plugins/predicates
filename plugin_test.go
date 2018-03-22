package main

import (
	"path/filepath"
	"testing"

	"github.com/zalando/skipper"
)

var pluginDir string = "./build"

func TestLoadPredicateGeoIP(t *testing.T) {
	if _, err := skipper.LoadPredicatePlugin(filepath.Join(pluginDir, "pred_geoip.so"), []string{"db=GeoLite2-Country.mmdb"}); err != nil {
		t.Errorf("failed to load plugin `geoip`: %s", err)
	}
}
