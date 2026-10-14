package mcpserver

import "testing"

func TestGetOnnxBinaryNameForPlatform(t *testing.T) {
	tests := map[string]struct {
		goos    string
		goarch  string
		want    string
		wantErr bool
	}{
		"windows-amd64": {goos: "windows", goarch: "amd64", want: "onnxruntime-win-x64-1.30.0.dll"},
		"darwin-arm64":  {goos: "darwin", goarch: "arm64", want: "onnxruntime-osx-arm64-1.30.0.dylib"},
		"darwin-amd64":  {goos: "darwin", goarch: "amd64", want: "onnxruntime-osx-x86_64-1.30.0.dylib"},
		"linux-amd64":   {goos: "linux", goarch: "amd64", want: "onnxruntime-linux-x64-1.30.0.so"},
		"linux-arm64":   {goos: "linux", goarch: "arm64", want: "onnxruntime-linux-aarch64-1.30.0.so"},
		"unsupported":   {goos: "freebsd", goarch: "amd64", wantErr: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := getOnnxBinaryNameForPlatform(tc.goos, tc.goarch)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for %s/%s, got %q", tc.goos, tc.goarch, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %s/%s: %v", tc.goos, tc.goarch, err)
			}
			if got != tc.want {
				t.Fatalf("for %s/%s got %q, want %q", tc.goos, tc.goarch, got, tc.want)
			}
		})
	}
}

func TestGetOnnxArchiveNameForPlatform(t *testing.T) {
	tests := map[string]struct {
		goos   string
		goarch string
		want   string
	}{
		"windows-amd64": {goos: "windows", goarch: "amd64", want: "onnxruntime-win-x64-1.30.0.zip"},
		"darwin-arm64":  {goos: "darwin", goarch: "arm64", want: "onnxruntime-osx-arm64-1.30.0.tgz"},
		"darwin-amd64":  {goos: "darwin", goarch: "amd64", want: "onnxruntime-osx-x86_64-1.30.0.tgz"},
		"linux-amd64":   {goos: "linux", goarch: "amd64", want: "onnxruntime-linux-x64-1.30.0.tgz"},
		"linux-arm64":   {goos: "linux", goarch: "arm64", want: "onnxruntime-linux-aarch64-1.30.0.tgz"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := getOnnxArchiveNameForPlatform(tc.goos, tc.goarch)
			if err != nil {
				t.Fatalf("unexpected error for %s/%s: %v", tc.goos, tc.goarch, err)
			}
			if got != tc.want {
				t.Fatalf("for %s/%s got %q, want %q", tc.goos, tc.goarch, got, tc.want)
			}
		})
	}
}
