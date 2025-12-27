//soppo:generated v1
package paths

import "os"
import "path/filepath"
import "runtime"

// SopmodDir returns the sopmod root directory (~/.sopmod).
// This is the base directory where all sopmod data is stored.
//
// ```sop,no_run
// import "fmt"
// fmt.Println(SopmodDir())
// // Output:
// // /home/user/.sopmod
// ```
func SopmodDir() string {
	home, _err0 := os.UserHomeDir()
	if _err0 != nil {
		err := _err0
		panic("could not determine home directory: " + err.Error())
	}

	return filepath.Join(home, ".sopmod")
}

// GoRoot returns the Go installations directory (~/.sopmod/go).
//
// ```sop,no_run
// import "fmt"
// fmt.Println(GoRoot())
// // Output:
// // /home/user/.sopmod/go
// ```
func GoRoot() string {
	return filepath.Join(SopmodDir(), "go")
}

// GoDir returns a specific Go version directory.
//
// ```sop,no_run
// import "fmt"
// fmt.Println(GoDir("1.22.0"))
// // Output:
// // /home/user/.sopmod/go/1.22.0
// ```
func GoDir(version string) string {
	return filepath.Join(GoRoot(), version)
}

// GoBinary returns the Go binary path for a specific version.
// On Windows, returns path ending in go.exe.
//
// ```sop,no_run
// import "fmt"
// fmt.Println(GoBinary("1.22.0"))
// // Output (Unix):
// // /home/user/.sopmod/go/1.22.0/go/bin/go
// ```
func GoBinary(version string) string {
	dir := GoDir(version)
	if runtime.GOOS == "windows" {
		return filepath.Join(dir, "go", "bin", "go.exe")
	}
	return filepath.Join(dir, "go", "bin", "go")
}

// SopRoot returns the sop installations directory (~/.sopmod/sop).
//
// ```sop,no_run
// import "fmt"
// fmt.Println(SopRoot())
// // Output:
// // /home/user/.sopmod/sop
// ```
func SopRoot() string {
	return filepath.Join(SopmodDir(), "sop")
}

// SopDir returns a specific sop version directory.
//
// ```sop,no_run
// import "fmt"
// fmt.Println(SopDir("0.5.0"))
// // Output:
// // /home/user/.sopmod/sop/0.5.0
// ```
func SopDir(version string) string {
	return filepath.Join(SopRoot(), version)
}

// SopBinary returns the sop binary path for a specific version.
// On Windows, returns path ending in sop.exe.
//
// ```sop,no_run
// import "fmt"
// fmt.Println(SopBinary("0.5.0"))
// // Output (Unix):
// // /home/user/.sopmod/sop/0.5.0/sop
// ```
func SopBinary(version string) string {
	dir := SopDir(version)
	if runtime.GOOS == "windows" {
		return filepath.Join(dir, "sop.exe")
	}
	return filepath.Join(dir, "sop")
}

// SoplsBinary returns the sopls binary path for a specific version.
// On Windows, returns path ending in sopls.exe.
//
// ```sop,no_run
// import "fmt"
// fmt.Println(SoplsBinary("0.5.0"))
// // Output (Unix):
// // /home/user/.sopmod/sop/0.5.0/sopls
// ```
func SoplsBinary(version string) string {
	dir := SopDir(version)
	if runtime.GOOS == "windows" {
		return filepath.Join(dir, "sopls.exe")
	}
	return filepath.Join(dir, "sopls")
}

// ConfigPath returns the config file path (~/.sopmod/config.toml).
//
// ```sop,no_run
// import "fmt"
// fmt.Println(ConfigPath())
// // Output:
// // /home/user/.sopmod/config.toml
// ```
func ConfigPath() string {
	return filepath.Join(SopmodDir(), "config.toml")
}

// BinDir returns the bin directory for the shim (~/.sopmod/bin).
// This directory should be added to PATH.
//
// ```sop,no_run
// import "fmt"
// fmt.Println(BinDir())
// // Output:
// // /home/user/.sopmod/bin
// ```
func BinDir() string {
	return filepath.Join(SopmodDir(), "bin")
}

// SopShim returns the shim binary path (~/.sopmod/bin/sop).
// On Windows, returns path ending in sop.exe.
//
// ```sop,no_run
// import "fmt"
// fmt.Println(SopShim())
// // Output (Unix):
// // /home/user/.sopmod/bin/sop
// ```
func SopShim() string {
	dir := BinDir()
	if runtime.GOOS == "windows" {
		return filepath.Join(dir, "sop.exe")
	}
	return filepath.Join(dir, "sop")
}

// SoplsShim returns the sopls shim binary path (~/.sopmod/bin/sopls).
// On Windows, returns path ending in sopls.exe.
//
// ```sop,no_run
// import "fmt"
// fmt.Println(SoplsShim())
// // Output (Unix):
// // /home/user/.sopmod/bin/sopls
// ```
func SoplsShim() string {
	dir := BinDir()
	if runtime.GOOS == "windows" {
		return filepath.Join(dir, "sopls.exe")
	}
	return filepath.Join(dir, "sopls")
}

// EnsureDirs creates the sopmod directory structure.
// Creates ~/.sopmod/go, ~/.sopmod/sop, and ~/.sopmod/bin.
//
// ```sop,no_run
// EnsureDirs() ? err {
// 	panic(err)
// }
// ```
func EnsureDirs() error {
	dirs := []string{GoRoot(), SopRoot(), BinDir()}
	for _, dir := range dirs {
		_err0 := os.MkdirAll(dir, 0o755)
		if _err0 != nil {
			return _err0
		}
	}
	return nil
}

