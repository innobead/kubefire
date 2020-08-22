package versionfinder

import (
	"github.com/innobead/kubefire/pkg/data"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
)

func TestKubeadmVersionFinder_GetLatestVersion(t *testing.T) {
	tests := []struct {
		name    string
		finder  *KubeadmVersionFinder
		want    *data.Version
		wantErr bool
	}{
		{
			"Correct version",
			NewKubeadmVersionFinder(),
			getKubeLatestVersionFromStableChannel(),
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.finder.GetLatestVersion()
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

func TestKubeadmVersionFinder_GetLatestMinorVersions(t *testing.T) {
	type args struct {
		afterVersion data.Version
	}
	tests := []struct {
		name             string
		finder           *KubeadmVersionFinder
		args             args
		wantVersionCount int
		wantErr          bool
	}{
		{
			"Correct version",
			NewKubeadmVersionFinder(),
			args{
				*getKubeLatestVersionFromStableChannel(),
			},
			3,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.finder.GetVersionsAfterVersion(tt.args.afterVersion)
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

func getKubeLatestVersionFromStableChannel() *data.Version {
	resp, err := http.Get("https://storage.googleapis.com/kubernetes-release/release/stable.txt")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return data.ParseVersion(string(body))
}

func TestKubeadmVersionFinder_GetCriToolMinorVersions(t *testing.T) {
	type args struct {
		afterVersion data.Version
	}
	tests := []struct {
		name             string
		finder           *KubeadmVersionFinder
		args             args
		wantVersionCount int
		wantErr          bool
	}{
		{
			"Correct version",
			NewKubeadmVersionFinder(),
			args{
				*getKubeLatestVersionFromStableChannel(),
			},
			3,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.finder.GetCritoolVersionsAfterVersion(tt.args.afterVersion)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCritoolVersionsAfterVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != tt.wantVersionCount {
				t.Errorf("GetCritoolVersionsAfterVersion() got len = %v, want len %v", len(got), tt.wantVersionCount)
			}
		})
	}
}
