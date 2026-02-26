package fs

import "testing"

func TestParsePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected PathInfo
	}{
		{
			name:     "local absolute path",
			input:    "/home/user/data",
			expected: PathInfo{Path: "/home/user/data"},
		},
		{
			name:     "local relative path",
			input:    "relative/path",
			expected: PathInfo{Path: "relative/path"},
		},
		{
			name:  "remote with user",
			input: "user@host:/remote/path",
			expected: PathInfo{
				IsRemote: true,
				User:     "user",
				Host:     "host",
				Path:     "/remote/path",
			},
		},
		{
			name:  "remote without user",
			input: "host:/remote/path",
			expected: PathInfo{
				IsRemote: true,
				Host:     "host",
				Path:     "/remote/path",
			},
		},
		{
			name:  "remote with domain host",
			input: "user@example.com:/data",
			expected: PathInfo{
				IsRemote: true,
				User:     "user",
				Host:     "example.com",
				Path:     "/data",
			},
		},
		{
			name:  "remote with IP",
			input: "root@192.168.1.1:/var/data",
			expected: PathInfo{
				IsRemote: true,
				User:     "root",
				Host:     "192.168.1.1",
				Path:     "/var/data",
			},
		},
		{
			name:     "windows drive letter",
			input:    "C:\\Users\\data",
			expected: PathInfo{Path: "C:\\Users\\data"},
		},
		{
			name:  "remote relative path",
			input: "user@host:relative/path",
			expected: PathInfo{
				IsRemote: true,
				User:     "user",
				Host:     "host",
				Path:     "relative/path",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParsePath(tt.input)
			if result != tt.expected {
				t.Errorf("ParsePath(%q) = %+v, want %+v", tt.input, result, tt.expected)
			}
		})
	}
}
