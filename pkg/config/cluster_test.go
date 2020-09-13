package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCluster_UpdateExtraOptions(t *testing.T) {
	tests := []struct {
		name         string
		extraOptions map[string]interface{}
		options      string
	}{
		{
			name: "valid extra options with comma separator",
			extraOptions: map[string]interface{}{
				"options": []string{
					"--k1=v1",
					"--k2=v2",
				},
			},
			options: `options="--k1=v1,--k2=v2"`,
		},
		{
			name: "valid many same extra options with comma separator",
			extraOptions: map[string]interface{}{
				"options": []string{
					"--k1=v1",
					"--k2=v2",
					"--k3=v3",
				},
			},
			options: `options="--k1=v1,--k2=v2" options="--k3=v3"`,
		},
		{
			name: "valid extra options with comma separator",
			extraOptions: map[string]interface{}{
				"options_1": []string{
					"--k1=v1",
					"--k2=v2",
				},
				"options_2": []string{
					"--k1=v1",
					"--k2=v2",
				},
			},
			options: `options_1="--k1=v1,--k2=v2" options_2='--k1=v1,--k2=v2'`,
		},
		{
			name: "valid extra options w/ single option",
			extraOptions: map[string]interface{}{
				"options": "v",
			},
			options: `options=v`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cluster{
				ExtraOptions: map[string]interface{}{},
			}
			c.UpdateExtraOptions(tt.options)

			assert.Equal(t, tt.extraOptions, c.ExtraOptions)
		})
	}
}

func TestCluster_ParseExtraOptions(t *testing.T) {
	type ExtraOptions struct {
		Options []string `json:"options"`
	}

	tests := []struct {
		name         string
		extraOptions map[string]interface{}
		options      interface{}
		wantErr      bool
	}{
		{
			name: "",
			extraOptions: map[string]interface{}{
				"options": []string{
					"--k1=v1",
				},
			},
			options: ExtraOptions{
				Options: []string{
					"--k1=v1",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cluster{
				ExtraOptions: tt.extraOptions,
			}

			options := ExtraOptions{}

			if err := c.ParseExtraOptions(&options); (err != nil) != tt.wantErr {
				t.Errorf("ParseExtraOptions() error = %v, wantErr %v", err, tt.wantErr)
			}

			assert.Equal(t, tt.options, options)
		})
	}
}
