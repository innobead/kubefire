//go:build !feature_run_on_ci
// +build !feature_run_on_ci

package versionfinder

import (
	"github.com/innobead/kubefire/pkg/data"
	"net/http"
	"path"
	"reflect"
	"strings"
	"testing"
)

func TestK3sVersionFinder_GetLatestVersion(t *testing.T) {
	tests := []struct {
		name    string
		want    *data.Version
		wantErr bool
	}{
		{
			"Correct version",
			data.ParseVersion(getK3sLatestVersionByRedirectUrl()),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &K3sVersionFinder{}
			got, err := k.GetLatestVersion()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLatestVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLatestVersion() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestK3sVersionFinder_GetLatestMinorVersions(t *testing.T) {
	tests := []struct {
		name             string
		wantVersionCount int
		wantErr          bool
	}{
		{
			"Last 3 supported version",
			3,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &K3sVersionFinder{}
			got, err := k.GetVersionsAfterVersion(data.Version{})
			if (err != nil) != tt.wantErr {
				t.Errorf("GetVersionsAfterVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != tt.wantVersionCount {
				t.Errorf("GetVersionsAfterVersion() got len = %v, want len %v", len(got), tt.wantVersionCount)
			}
		})
	}
}

func getK3sLatestVersionByRedirectUrl() string {
	get, _ := http.Get("https://update.k3s.io/v1-release/channels/latest")
	return strings.ReplaceAll(path.Base(get.Request.URL.Path), "+k3s1", "")
}
