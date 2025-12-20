//soppo:generated
package compat

import "testing"

func TestParseVersion(t *testing.T) {
	tests := []struct {
		input   string
		major   int
		minor   int
		patch   int
		wantErr bool
	}{
		{input: "1.22.0", major: 1, minor: 22, patch: 0, wantErr: false},
		{input: "1.21", major: 1, minor: 21, patch: 0, wantErr: false},
		{input: "2", major: 2, minor: 0, patch: 0, wantErr: false},
		{input: "0.5.0", major: 0, minor: 5, patch: 0, wantErr: false},
		{input: "invalid", major: 0, minor: 0, patch: 0, wantErr: true},
	}

	for _, tt := range tests {
		major, minor, patch, err := parseVersion(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("parseVersion(%q) expected error, got nil", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseVersion(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if major != tt.major || minor != tt.minor || patch != tt.patch {
			t.Errorf("parseVersion(%q) = (%d, %d, %d), want (%d, %d, %d)", tt.input, major, minor, patch, tt.major, tt.minor, tt.patch)
		}
	}
}

func TestVersionAtLeast(t *testing.T) {
	tests := []struct {
		aMajor int
		aMinor int
		aPatch int
		bMajor int
		bMinor int
		bPatch int
		want   bool
	}{
		{aMajor: 1, aMinor: 22, aPatch: 0, bMajor: 1, bMinor: 21, bPatch: 0, want: true},
		{aMajor: 1, aMinor: 21, aPatch: 0, bMajor: 1, bMinor: 21, bPatch: 0, want: true},
		{aMajor: 1, aMinor: 20, aPatch: 0, bMajor: 1, bMinor: 21, bPatch: 0, want: false},
		{aMajor: 2, aMinor: 0, aPatch: 0, bMajor: 1, bMinor: 99, bPatch: 0, want: true},
		{aMajor: 1, aMinor: 21, aPatch: 5, bMajor: 1, bMinor: 21, bPatch: 3, want: true},
		{aMajor: 1, aMinor: 21, aPatch: 3, bMajor: 1, bMinor: 21, bPatch: 5, want: false},
	}

	// 1.22.0 >= 1.21.0
	// 1.21.0 >= 1.21.0
	// 1.20.0 < 1.21.0
	// 2.0.0 >= 1.99.0
	// 1.21.5 >= 1.21.3
	// 1.21.3 < 1.21.5
	for _, tt := range tests {
		got := versionAtLeast(tt.aMajor, tt.aMinor, tt.aPatch, tt.bMajor, tt.bMinor, tt.bPatch)
		if got != tt.want {
			t.Errorf("versionAtLeast(%d.%d.%d, %d.%d.%d) = %v, want %v", tt.aMajor, tt.aMinor, tt.aPatch, tt.bMajor, tt.bMinor, tt.bPatch, got, tt.want)
		}
	}
}

func TestGoCompatFor(t *testing.T) {
	tests := []struct {
		sopVersion string
		wantMin    string
		wantNil    bool
	}{
		{sopVersion: "0.5.0", wantMin: "1.21", wantNil: false},
		{sopVersion: "v0.5.0", wantMin: "1.21", wantNil: false},
		{sopVersion: "0.1.0", wantMin: "1.21", wantNil: false},
		{sopVersion: "1.0.0", wantMin: "1.21", wantNil: false},
		{sopVersion: "0.0.1", wantMin: "1.21", wantNil: false},
	}

	for _, tt := range tests {
		got := GoCompatFor(tt.sopVersion)
		if tt.wantNil {
			if got != nil {
				t.Errorf("GoCompatFor(%q) = %v, want nil", tt.sopVersion, got)
			}
			continue
		}
		if got == nil {
			t.Errorf("GoCompatFor(%q) = nil, want non-nil", tt.sopVersion)
			continue
		}
		if got.Min != tt.wantMin {
			t.Errorf("GoCompatFor(%q).Min = %q, want %q", tt.sopVersion, got.Min, tt.wantMin)
		}
	}
}

func TestIsGoCompatible(t *testing.T) {
	tests := []struct {
		goVersion  string
		sopVersion string
		want       bool
	}{
		{goVersion: "1.22.0", sopVersion: "0.5.0", want: true},
		{goVersion: "1.21.0", sopVersion: "0.5.0", want: true},
		{goVersion: "1.21", sopVersion: "0.5.0", want: true},
		{goVersion: "1.20.0", sopVersion: "0.5.0", want: false},
		{goVersion: "1.19.0", sopVersion: "0.5.0", want: false},
		{goVersion: "2.0.0", sopVersion: "0.5.0", want: true},
	}

	for _, tt := range tests {
		got := IsGoCompatible(tt.goVersion, tt.sopVersion)
		if got != tt.want {
			t.Errorf("IsGoCompatible(%q, %q) = %v, want %v", tt.goVersion, tt.sopVersion, got, tt.want)
		}
	}
}

func TestCompatMessage(t *testing.T) {
	tests := []struct {
		sopVersion string
		want       string
	}{
		{sopVersion: "0.5.0", want: "sop 0.5.0 requires go 1.21 or later"},
		{sopVersion: "invalid", want: "sop invalid has unknown go requirements"},
	}

	for _, tt := range tests {
		got := CompatMessage(tt.sopVersion)
		if got != tt.want {
			t.Errorf("CompatMessage(%q) = %q, want %q", tt.sopVersion, got, tt.want)
		}
	}
}

