// Copyright (c) Jiaqi Liu
// SPDX-License-Identifier: MPL-2.0

package ssl

import "testing"

func TestWriteToFile(t *testing.T) {
	filename1, err := WriteToFile("foo")
	if err != nil {
		t.Error(err)
	}

	filename2, err := WriteToFile("foo")
	if err != nil {
		t.Error(err)
	}

	t.Logf("filename 1: %s; filename 2: %s", filename1, filename2)
	if filename1 == filename2 {
		t.Errorf("WritingToFile on the same content should generate different file names across invocations. But 2 function calls generate a same filename of %s", filename1)
	}
}

func TestGetHomeDir(t *testing.T) {
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
