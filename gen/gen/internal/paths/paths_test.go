//soppo:generated
package paths

import "runtime"
import "strings"
import "testing"

func TestSopmodDir(t *testing.T) {
	dir := SopmodDir()
	if (!strings.HasSuffix(dir, ".sopmod")) {
		t.Errorf("SopmodDir() = %q, want suffix .sopmod", dir)
	}
}

func TestGoRoot(t *testing.T) {
	root := GoRoot()
	if (!strings.HasSuffix(root, ".sopmod/go")) && (!strings.HasSuffix(root, ".sopmod\\go")) {
		t.Errorf("GoRoot() = %q, want suffix .sopmod/go", root)
	}
}

func TestGoDir(t *testing.T) {
	tests := []struct {
		version string
		suffix  string
	}{
		{version: "1.22.0", suffix: "go/1.22.0"},
		{version: "1.21", suffix: "go/1.21"},
		{version: "2.0.0", suffix: "go/2.0.0"},
	}

	for _, tt := range tests {
		got := GoDir(tt.version)

		// Handle both Unix and Windows path separators
		wantUnix := ".sopmod/" + tt.suffix
		wantWin := ".sopmod\\" + strings.ReplaceAll(tt.suffix, "/", "\\")
		if (!strings.HasSuffix(got, wantUnix)) && (!strings.HasSuffix(got, wantWin)) {
			t.Errorf("GoDir(%q) = %q, want suffix %q", tt.version, got, wantUnix)
		}
	}
}

func TestGoBinary(t *testing.T) {
	version := "1.22.0"
	got := GoBinary(version)

	// Check version is in the path
	if (!strings.Contains(got, version)) {
		t.Errorf("GoBinary(%q) = %q, should contain version", version, got)
	}

	// Check correct binary name for platform
	if runtime.GOOS == "windows" {
		if (!strings.HasSuffix(got, "go.exe")) {
			t.Errorf("GoBinary(%q) = %q, want suffix go.exe on Windows", version, got)
		}
	} else {
		if (!strings.HasSuffix(got, "/go")) {
			t.Errorf("GoBinary(%q) = %q, want suffix /go on Unix", version, got)
		}
	}

	// Check the directory structure go/bin/go
	if (!strings.Contains(got, "go/bin/go")) && (!strings.Contains(got, "go\\bin\\go")) {
		t.Errorf("GoBinary(%q) = %q, should contain go/bin/go structure", version, got)
	}
}

func TestSopRoot(t *testing.T) {
	root := SopRoot()
	if (!strings.HasSuffix(root, ".sopmod/sop")) && (!strings.HasSuffix(root, ".sopmod\\sop")) {
		t.Errorf("SopRoot() = %q, want suffix .sopmod/sop", root)
	}
}

func TestSopDir(t *testing.T) {
	tests := []struct {
		version string
		suffix  string
	}{
		{version: "0.5.0", suffix: "sop/0.5.0"},
		{version: "0.1.0", suffix: "sop/0.1.0"},
		{version: "1.0.0", suffix: "sop/1.0.0"},
	}

	for _, tt := range tests {
		got := SopDir(tt.version)
		wantUnix := ".sopmod/" + tt.suffix
		wantWin := ".sopmod\\" + strings.ReplaceAll(tt.suffix, "/", "\\")
		if (!strings.HasSuffix(got, wantUnix)) && (!strings.HasSuffix(got, wantWin)) {
			t.Errorf("SopDir(%q) = %q, want suffix %q", tt.version, got, wantUnix)
		}
	}
}

func TestSopBinary(t *testing.T) {
	version := "0.5.0"
	got := SopBinary(version)

	// Check version is in the path
	if (!strings.Contains(got, version)) {
		t.Errorf("SopBinary(%q) = %q, should contain version", version, got)
	}

	// Check correct binary name for platform
	if runtime.GOOS == "windows" {
		if (!strings.HasSuffix(got, "sop.exe")) {
			t.Errorf("SopBinary(%q) = %q, want suffix sop.exe on Windows", version, got)
		}
	} else {
		if (!strings.HasSuffix(got, "/sop")) {
			t.Errorf("SopBinary(%q) = %q, want suffix /sop on Unix", version, got)
		}
	}
}

func TestConfigPath(t *testing.T) {
	got := ConfigPath()
	if (!strings.HasSuffix(got, "config.toml")) {
		t.Errorf("ConfigPath() = %q, want suffix config.toml", got)
	}
	if (!strings.Contains(got, ".sopmod")) {
		t.Errorf("ConfigPath() = %q, should contain .sopmod", got)
	}
}

func TestBinDir(t *testing.T) {
	got := BinDir()
	if (!strings.HasSuffix(got, ".sopmod/bin")) && (!strings.HasSuffix(got, ".sopmod\\bin")) {
		t.Errorf("BinDir() = %q, want suffix .sopmod/bin", got)
	}
}

func TestSopShim(t *testing.T) {
	got := SopShim()

	// Check it's in the bin directory
	if (!strings.Contains(got, "bin")) {
		t.Errorf("SopShim() = %q, should contain bin", got)
	}

	// Check correct binary name for platform
	if runtime.GOOS == "windows" {
		if (!strings.HasSuffix(got, "sop.exe")) {
			t.Errorf("SopShim() = %q, want suffix sop.exe on Windows", got)
		}
	} else {
		if (!strings.HasSuffix(got, "/sop")) {
			t.Errorf("SopShim() = %q, want suffix /sop on Unix", got)
		}
	}
}

func TestPathConsistency(t *testing.T) {
	// Test that paths are consistent with each other
	sopmodDir := SopmodDir()

	// GoRoot should be under SopmodDir
	goRoot := GoRoot()
	if (!strings.HasPrefix(goRoot, sopmodDir)) {
		t.Errorf("GoRoot() = %q should be under SopmodDir() = %q", goRoot, sopmodDir)
	}

	// SopRoot should be under SopmodDir
	sopRoot := SopRoot()
	if (!strings.HasPrefix(sopRoot, sopmodDir)) {
		t.Errorf("SopRoot() = %q should be under SopmodDir() = %q", sopRoot, sopmodDir)
	}

	// BinDir should be under SopmodDir
	binDir := BinDir()
	if (!strings.HasPrefix(binDir, sopmodDir)) {
		t.Errorf("BinDir() = %q should be under SopmodDir() = %q", binDir, sopmodDir)
	}

	// ConfigPath should be under SopmodDir
	configPath := ConfigPath()
	if (!strings.HasPrefix(configPath, sopmodDir)) {
		t.Errorf("ConfigPath() = %q should be under SopmodDir() = %q", configPath, sopmodDir)
	}

	// GoDir should be under GoRoot
	goDir := GoDir("1.22.0")
	if (!strings.HasPrefix(goDir, goRoot)) {
		t.Errorf("GoDir(1.22.0) = %q should be under GoRoot() = %q", goDir, goRoot)
	}

	// SopDir should be under SopRoot
	sopDir := SopDir("0.5.0")
	if (!strings.HasPrefix(sopDir, sopRoot)) {
		t.Errorf("SopDir(0.5.0) = %q should be under SopRoot() = %q", sopDir, sopRoot)
	}

	// SopShim should be under BinDir
	sopShim := SopShim()
	if (!strings.HasPrefix(sopShim, binDir)) {
		t.Errorf("SopShim() = %q should be under BinDir() = %q", sopShim, binDir)
	}
}

