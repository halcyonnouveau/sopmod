//soppo:generated v1
package shim

import "fmt"
import "os"
import "path/filepath"
import "strings"
import "syscall"
import "github.com/halcyonnouveau/sopmod/gen/internal/config"
import "github.com/halcyonnouveau/sopmod/gen/internal/install"
import "github.com/halcyonnouveau/sopmod/gen/internal/paths"

// Run executes the sop shim, resolving versions and setting up the environment
func Run() error {
	return runBinary(paths.SopBinary)
}

// RunLsp executes the sopls shim, resolving versions and setting up the environment
func RunLsp() error {
	return runBinary(paths.SoplsBinary)
}

func runBinary(binaryPathFn func(string) string) error {
	wantedSop, _err0 := findSopVersion()
	if _err0 != nil {
		return _err0
	}
	sopVersion := ResolveInstalledVersion(wantedSop, install.ListInstalledSop())
	if sopVersion == "" {
		return fmt.Errorf("sop %s is not installed. Run `sopmod install sop %s`", wantedSop, wantedSop)
	}
	binary := binaryPathFn(sopVersion)

	// Set up environment with managed Go version
	env := os.Environ()
	if wantedGo, err := findGoVersion(); err == nil && wantedGo != "" {
		goVersion := ResolveInstalledVersion(wantedGo, install.ListInstalledGo())
		if goVersion == "" {
			return fmt.Errorf("go %s is not installed. Run `sopmod install go %s`", wantedGo, wantedGo)
		}
		goBinDir := filepath.Dir(paths.GoBinary(goVersion))
		env = append(env, "PATH=" + goBinDir + ":" + os.Getenv("PATH"))
	}

	// Exec binary with all original args
	args := append([]string{binary}, os.Args[1:]...)
	return syscall.Exec(binary, args, env)
}

// Install copies the current binary to both sop and sopls shim locations
func Install() error {
	currentExe, _err0 := os.Executable()
	if _err0 != nil {
		return _err0
	}

	// Install sop shim
	sopShimPath := paths.SopShim()
	os.Remove(sopShimPath)
	_err1 := copyFile(currentExe, sopShimPath)
	if _err1 != nil {
		return _err1
	}
	_err2 := os.Chmod(sopShimPath, 0o755)
	if _err2 != nil {
		return _err2
	}

	// Install sopls shim
	soplsShimPath := paths.SoplsShim()
	os.Remove(soplsShimPath)
	_err3 := copyFile(currentExe, soplsShimPath)
	if _err3 != nil {
		return _err3
	}
	return os.Chmod(soplsShimPath, 0o755)
}

func findSopVersion() (string, error) {
	// Check sop.mod in current dir and parents
	if projectCfg, err := findProjectConfig(); err == nil && projectCfg != nil && projectCfg.Sop != nil {
		return (*projectCfg.Sop), nil
	}

	// Fall back to default
	cfg := config.Load()
	if cfg.DefaultSop != nil {
		return (*cfg.DefaultSop), nil
	}

	return "", fmt.Errorf("no sop version configured. Run `sopmod install sop latest`")
}

func findGoVersion() (string, error) {
	// Check sop.mod in current dir and parents
	if projectCfg, err := findProjectConfig(); err == nil && projectCfg != nil && projectCfg.Go != nil {
		return (*projectCfg.Go), nil
	}

	// Fall back to default
	cfg := config.Load()
	if cfg.DefaultGo != nil {
		return (*cfg.DefaultGo), nil
	}

	return "", nil
}

// findProjectConfig walks up the directory tree looking for sop.mod
//soppo:nilable : 0
func findProjectConfig() (*config.ProjectConfig, error) {
	current, _err0 := os.Getwd()
	if _err0 != nil {
		return nil, _err0
	}

	for {
		sopModPath := filepath.Join(current, "sop.mod")
		if fileExists(sopModPath) {
			return config.LoadProjectConfig(current)
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	return nil, nil
}

// ResolveInstalledVersion finds the best installed version matching a version or prefix.
// Returns exact match if found, otherwise highest version matching the prefix.
//
// ```sop
// import "fmt"
// // Exact match
// fmt.Println(ResolveInstalledVersion("1.22.0", []string{"1.21.0", "1.22.0", "1.23.0"}))
// // Prefix match - returns highest
// fmt.Println(ResolveInstalledVersion("1.22", []string{"1.22.0", "1.22.5", "1.23.0"}))
// // Output:
// // 1.22.0
// // 1.22.5
// ```
func ResolveInstalledVersion(wanted string, installed []string) string {
	// Exact match first
	for _, v := range installed {
		if v == wanted {
			return v
		}
	}

	// Prefix match - find highest matching version
	prefix := wanted + "."
	var best string
	for _, v := range installed {
		if strings.HasPrefix(v, prefix) {
			if best == "" || CompareVersions(v, best) > 0 {
				best = v
			}
		}
	}

	return best
}

// CompareVersions compares two semver strings.
// Returns negative if a < b, zero if a == b, positive if a > b.
//
// ```sop
// import "fmt"
// fmt.Println(CompareVersions("1.22.0", "1.21.0") > 0)
// fmt.Println(CompareVersions("1.21.0", "1.22.0") < 0)
// fmt.Println(CompareVersions("1.22.0", "1.22.0") == 0)
// // Output:
// // true
// // true
// // true
// ```
func CompareVersions(a string, b string) int {
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")

	for i := 0; i < len(aParts) && i < len(bParts); i++ {
		var aNum, bNum int
		fmt.Sscanf(aParts[i], "%d", (&aNum))
		fmt.Sscanf(bParts[i], "%d", (&bNum))
		if aNum != bNum {
			return aNum - bNum
		}
	}
	return len(aParts) - len(bParts)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && (!info.IsDir())
}

func copyFile(src string, dst string) error {
	data, _err0 := os.ReadFile(src)
	if _err0 != nil {
		return _err0
	}
	return os.WriteFile(dst, data, 0o755)
}

