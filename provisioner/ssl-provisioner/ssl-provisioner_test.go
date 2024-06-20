// Copyright (c) Jiaqi Liu
// SPDX-License-Identifier: MPL-2.0

package sslProvisioner

import "testing"

func Test_getHomeDir(t *testing.T) {
	data := []struct {
		name        string
		configValue string
		expected    string
	}{
		{"regular directory is specified", "/", "/"},
		{"no directory is specified as home dir", "", "/home/ubuntu"},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			actual := GetHomeDir(d.configValue)
			if actual != d.expected {
				t.Errorf("Expected %s, got %s", d.expected, actual)
			}
		})
	}
}
