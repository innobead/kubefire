package config

import "testing"

func TestIsReleasedTagVersion(t *testing.T) {
	type args struct {
		version string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"tag from release", args{"v1.0.1"}, true},
		{"tag not from release", args{"v1.0.1-dirty"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsReleasedTagVersion(tt.args.version); got != tt.want {
				t.Errorf("IsReleasedTagVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetTagVersionForDownloadScript(t *testing.T) {
	type args struct {
		version string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"tag from release", args{"v1.0.1"}, "v1.0.1"},
		{"tag not from release", args{"v1.0.1-dirty"}, "master"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetTagVersionForDownloadScript(tt.args.version); got != tt.want {
				t.Errorf("GetTagVersionForDownloadScript() = %v, want %v", got, tt.want)
			}
		})
	}
}
