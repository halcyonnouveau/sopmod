//soppo:generated v1
package install

import "archive/tar"
import "archive/zip"
import "compress/gzip"
import "encoding/json"
import "errors"
import "fmt"
import "io"
import "net/http"
import "os"
import "path/filepath"
import "runtime"
import "strings"
import "github.com/halcyonnouveau/sopmod/gen/internal/compat"
import "github.com/halcyonnouveau/sopmod/gen/internal/paths"

// Platform holds OS and architecture info for downloads
type Platform struct {
	OS string
	Arch string
}

// DetectPlatform returns the current platform
//soppo:nilable : 0
func DetectPlatform() (*Platform, error) {
	var osName string
	switch runtime.GOOS {
	case "linux":
		osName = "linux"
	case "darwin":
		osName = "darwin"
	case "windows":
		osName = "windows"
	default:
		return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	var arch string
	switch runtime.GOARCH {
	case "amd64":
		arch = "amd64"
	case "arm64":
		arch = "arm64"
	default:
		return nil, fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
	}

	return (&Platform{OS: osName, Arch: arch}), nil
}

func (p *Platform) GoArchiveExt() string {
	if p.OS == "windows" {
		return "zip"
	}
	return "tar.gz"
}

// GoRelease represents a Go release from go.dev/dl/?mode=json
type GoRelease struct {
	Version string `json:"version"`
	Stable bool `json:"stable"`
}

// ResolveLatestGo resolves "latest" to the actual latest stable Go version
func ResolveLatestGo() (string, error) {
	resp, _err0 := http.Get("https://go.dev/dl/?mode=json")
	if _err0 != nil {
		return "", _err0
	}
	defer resp.Body.Close()

	releases := []GoRelease{}
	_err1 := json.NewDecoder(resp.Body).Decode((&releases))
	if _err1 != nil {
		return "", _err1
	}

	for _, r := range releases {
		if r.Stable {
			return strings.TrimPrefix(r.Version, "go"), nil
		}
	}

	return "", errors.New("no stable Go version found")
}

// ResolveGoVersion resolves a Go version, handling "latest" and partial versions
func ResolveGoVersion(version string) (string, error) {
	if version == "latest" {
		return ResolveLatestGo()
	}

	// Check if it's a partial version like "1.22"
	parts := strings.Split(version, ".")
	if len(parts) == 2 {
		resp, _err0 := http.Get("https://go.dev/dl/?mode=json")
		if _err0 != nil {
			return "", _err0
		}
		defer resp.Body.Close()

		releases := []GoRelease{}
		_err1 := json.NewDecoder(resp.Body).Decode((&releases))
		if _err1 != nil {
			return "", _err1
		}

		prefix := "go" + version
		for _, r := range releases {
			if r.Stable && strings.HasPrefix(r.Version, prefix) {
				return strings.TrimPrefix(r.Version, "go"), nil
			}
		}
	}

	return version, nil
}

// InstallGo installs a specific Go version
func InstallGo(version string, verbose bool) (string, error) {
	resolved, _err0 := ResolveGoVersion(version)
	if _err0 != nil {
		return "", _err0
	}
	platform, _err1 := DetectPlatform()
	if _err1 != nil {
		return "", _err1
	}

	dest := paths.GoDir(resolved)
	if dirExists(dest) {
		fmt.Printf("\033[32m✓\033[0m go \033[1m%s\033[0m is already installed\n", resolved)
		return resolved, nil
	}

	ext := platform.GoArchiveExt()
	filename := fmt.Sprintf("go%s.%s-%s.%s", resolved, platform.OS, platform.Arch, ext)
	url := "https://go.dev/dl/" + filename

	if verbose {
		fmt.Printf("Downloading go %s from %s\n", resolved, url)
	}

	// Download
	resp, _err2 := http.Get(url)
	if _err2 != nil {
		return "", _err2
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("version not found: go %s for %s-%s", resolved, platform.OS, platform.Arch)
	}

	// Create temp file
	tmpFile, _err3 := os.CreateTemp("", "go-*." + ext)
	if _err3 != nil {
		return "", _err3
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Download with progress
	fmt.Printf("Downloading go %s\n", resolved)
	_, _err4 := io.Copy(tmpFile, resp.Body)
	if _err4 != nil {
		return "", _err4
	}
	tmpFile.Close()
	fmt.Println("Download complete")

	// Extract
	if verbose {
		fmt.Printf("Extracting to %s\n", dest)
	}

	_err5 := os.MkdirAll(dest, 0o755)
	if _err5 != nil {
		return "", _err5
	}

	if ext == "zip" {
		_err6 := extractZip(tmpFile.Name(), dest)
		if _err6 != nil {
			return "", _err6
		}
	} else {
		_err7 := extractTarGz(tmpFile.Name(), dest)
		if _err7 != nil {
			return "", _err7
		}
	}

	// Verify
	goBin := paths.GoBinary(resolved)
	if (!fileExists(goBin)) {
		return "", fmt.Errorf("go binary not found at %s", goBin)
	}

	fmt.Printf("\033[32m✓\033[0m go \033[1m%s\033[0m installed successfully\n", resolved)
	return resolved, nil
}

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets []GitHubAsset `json:"assets"`
}

// GitHubAsset represents a GitHub release asset
type GitHubAsset struct {
	Name string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// ResolveLatestSop resolves "latest" to the actual latest sop version
func ResolveLatestSop() (string, error) {
	req, _err0 := http.NewRequest("GET", "https://api.github.com/repos/halcyonnouveau/soppo/releases/latest", nil)
	if _err0 != nil {
		return "", _err0
	}
	req.Header.Set("User-Agent", "sopmod")

	resp, _err1 := http.DefaultClient.Do(req)
	if _err1 != nil {
		return "", _err1
	}
	defer resp.Body.Close()

	var release GitHubRelease
	_err2 := json.NewDecoder(resp.Body).Decode((&release))
	if _err2 != nil {
		return "", _err2
	}

	return strings.TrimPrefix(release.TagName, "v"), nil
}

// ResolveSopVersion resolves a sop version, handling "latest"
func ResolveSopVersion(version string) (string, error) {
	if version == "latest" {
		return ResolveLatestSop()
	}
	return version, nil
}

// InstallSop installs a specific sop version
func InstallSop(version string, verbose bool) (string, error) {
	resolved, _err0 := ResolveSopVersion(version)
	if _err0 != nil {
		return "", _err0
	}
	platform, _err1 := DetectPlatform()
	if _err1 != nil {
		return "", _err1
	}

	dest := paths.SopDir(resolved)
	if dirExists(dest) {
		fmt.Printf("\033[32m✓\033[0m sop \033[1m%s\033[0m is already installed\n", resolved)
		return resolved, nil
	}

	// Check Go compatibility
	compatInfo := compat.GoCompatFor(resolved)
	if compatInfo != nil {
		installedGo := ListInstalledGo()
		hasCompatible := false
		for _, v := range installedGo {
			if compat.IsGoCompatible(v, resolved) {
				hasCompatible = true
				break
			}
		}
		if (!hasCompatible) && len(installedGo) > 0 {
			fmt.Printf("\033[33mwarning:\033[0m %s\n", compat.CompatMessage(resolved))
			fmt.Printf("  Installed go versions: %s\n", strings.Join(installedGo, ", "))
			fmt.Printf("  \033[36mhint:\033[0m run \033[1msopmod install go %s\033[0m\n", compatInfo.Min)
		}
	}

	// Map platform to Rust target triple
	var targetTriple string
	switch platform.OS + "-" + platform.Arch {
	case "linux-amd64":
		targetTriple = "x86_64-unknown-linux-gnu"
	case "linux-arm64":
		targetTriple = "aarch64-unknown-linux-gnu"
	case "darwin-amd64":
		targetTriple = "x86_64-apple-darwin"
	case "darwin-arm64":
		targetTriple = "aarch64-apple-darwin"
	case "windows-amd64":
		targetTriple = "x86_64-pc-windows-msvc"
	case "windows-arm64":
		targetTriple = "aarch64-pc-windows-msvc"
	default:
		return "", fmt.Errorf("unsupported platform: %s-%s", platform.OS, platform.Arch)
	}

	// Fetch release info
	tag := "v" + resolved
	releaseURL := fmt.Sprintf("https://api.github.com/repos/halcyonnouveau/soppo/releases/tags/%s", tag)

	req, _err2 := http.NewRequest("GET", releaseURL, nil)
	if _err2 != nil {
		return "", _err2
	}
	req.Header.Set("User-Agent", "sopmod")

	resp, _err3 := http.DefaultClient.Do(req)
	if _err3 != nil {
		return "", _err3
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("version not found: sop %s", resolved)
	}

	var release GitHubRelease
	_err4 := json.NewDecoder(resp.Body).Decode((&release))
	if _err4 != nil {
		return "", _err4
	}

	// Find the sop and sopls assets
	sopPattern := "sop-" + targetTriple
	soplsPattern := "sopls-" + targetTriple
	var sopAsset, soplsAsset *GitHubAsset
	for i := range release.Assets {
		name := release.Assets[i].Name
		if sopAsset == nil && strings.HasPrefix(name, sopPattern) {
			sopAsset = (&release.Assets[i])
		}
		if soplsAsset == nil && strings.HasPrefix(name, soplsPattern) {
			soplsAsset = (&release.Assets[i])
		}
	}
	if sopAsset == nil {
		return "", fmt.Errorf("version not found: sop %s for %s", resolved, targetTriple)
	}

	_err5 := os.MkdirAll(dest, 0o755)
	if _err5 != nil {
		return "", _err5
	}

	// Download sop
	_err6 := downloadBinary(sopAsset, dest, "sop", verbose)
	if _err6 != nil {
		return "", _err6
	}

	// Download sopls if available
	if soplsAsset != nil {
		_err7 := downloadBinary(soplsAsset, dest, "sopls", verbose)
		if _err7 != nil {
			return "", _err7
		}
	}

	// Verify sop
	sopBin := paths.SopBinary(resolved)
	if (!fileExists(sopBin)) {
		return "", fmt.Errorf("sop binary not found at %s", sopBin)
	}

	fmt.Printf("\033[32m✓\033[0m sop \033[1m%s\033[0m installed successfully\n", resolved)
	return resolved, nil
}

// ListInstalledGo returns a list of installed Go versions
func ListInstalledGo() []string {
	goRoot := paths.GoRoot()
	if (!dirExists(goRoot)) {
		return []string{}
	}

	entries, _err0 := os.ReadDir(goRoot)
	if _err0 != nil {
		return nil
	}

	versions := []string{}
	for _, e := range entries {
		if e.IsDir() {
			versions = append(versions, e.Name())
		}
	}
	return versions
}

// ListInstalledSop returns a list of installed sop versions
func ListInstalledSop() []string {
	sopRoot := paths.SopRoot()
	if (!dirExists(sopRoot)) {
		return []string{}
	}

	entries, _err0 := os.ReadDir(sopRoot)
	if _err0 != nil {
		return nil
	}

	versions := []string{}
	for _, e := range entries {
		if e.IsDir() {
			versions = append(versions, e.Name())
		}
	}
	return versions
}

// RemoveGo removes an installed Go version
func RemoveGo(version string) error {
	dir := paths.GoDir(version)
	if (!dirExists(dir)) {
		return fmt.Errorf("version not found: go %s", version)
	}
	_err0 := os.RemoveAll(dir)
	if _err0 != nil {
		return _err0
	}
	fmt.Printf("\033[32m✓\033[0m Removed go \033[1m%s\033[0m\n", version)
	return nil
}

// RemoveSop removes an installed sop version
func RemoveSop(version string) error {
	dir := paths.SopDir(version)
	if (!dirExists(dir)) {
		return fmt.Errorf("version not found: sop %s", version)
	}
	_err0 := os.RemoveAll(dir)
	if _err0 != nil {
		return _err0
	}
	fmt.Printf("\033[32m✓\033[0m Removed sop \033[1m%s\033[0m\n", version)
	return nil
}

func extractTarGz(archivePath string, dest string) error {
	f, _err0 := os.Open(archivePath)
	if _err0 != nil {
		return _err0
	}
	defer f.Close()

	gzr, _err1 := gzip.NewReader(f)
	if _err1 != nil {
		return _err1
	}
	defer gzr.Close()
	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			_err2 := os.MkdirAll(target, 0o755)
			if _err2 != nil {
				return _err2
			}
		case tar.TypeReg:
			_err3 := os.MkdirAll(filepath.Dir(target), 0o755)
			if _err3 != nil {
				return _err3
			}
			outFile, _err4 := os.Create(target)
			if _err4 != nil {
				return _err4
			}
			_, err = io.Copy(outFile, tr)
			outFile.Close()
			if err != nil {
				return err
			}
			os.Chmod(target, os.FileMode(header.Mode))
		}
	}
	return nil
}

func extractZip(archivePath string, dest string) error {
	r, _err0 := zip.OpenReader(archivePath)
	if _err0 != nil {
		return _err0
	}
	defer r.Close()

	for _, f := range r.File {
		if f == nil {
			continue
		}

		target := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(target, 0o755)
			continue
		}

		_err1 := os.MkdirAll(filepath.Dir(target), 0o755)
		if _err1 != nil {
			return _err1
		}

		outFile, _err2 := os.Create(target)
		if _err2 != nil {
			return _err2
		}
		rc, _err3 := f.Open()
		if _err3 != nil {
			return _err3
		}

		_, err := io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			return err
		}

		os.Chmod(target, f.Mode())
	}
	return nil
}

func copyFile(src string, dst string) error {
	in, _err0 := os.Open(src)
	if _err0 != nil {
		return _err0
	}
	defer in.Close()

	out, _err1 := os.Create(dst)
	if _err1 != nil {
		return _err1
	}
	defer out.Close()

	_, err := io.Copy(out, in)
	return err
}

func downloadBinary(asset *GitHubAsset, dest string, name string, verbose bool) error {
	if verbose {
		fmt.Printf("Downloading %s from %s\n", name, asset.BrowserDownloadURL)
	}

	req, _err0 := http.NewRequest("GET", asset.BrowserDownloadURL, nil)
	if _err0 != nil {
		return _err0
	}
	req.Header.Set("User-Agent", "sopmod")

	resp, _err1 := http.DefaultClient.Do(req)
	if _err1 != nil {
		return _err1
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download %s: %s", name, resp.Status)
	}

	// Create temp file
	tmpFile, _err2 := os.CreateTemp("", name + "-*")
	if _err2 != nil {
		return _err2
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	fmt.Printf("Downloading %s\n", name)
	_, _err3 := io.Copy(tmpFile, resp.Body)
	if _err3 != nil {
		return _err3
	}
	tmpFile.Close()

	// Extract or copy
	if strings.HasSuffix(asset.Name, ".zip") {
		_err4 := extractZip(tmpFile.Name(), dest)
		if _err4 != nil {
			return _err4
		}
	} else {
		if strings.HasSuffix(asset.Name, ".tar.gz") || strings.HasSuffix(asset.Name, ".tgz") {
			_err5 := extractTarGz(tmpFile.Name(), dest)
			if _err5 != nil {
				return _err5
			}
		} else {
			// Raw binary
			binaryName := name
			if runtime.GOOS == "windows" {
				binaryName = name + ".exe"
			}
			destPath := filepath.Join(dest, binaryName)
			_err6 := copyFile(tmpFile.Name(), destPath)
			if _err6 != nil {
				return _err6
			}
			os.Chmod(destPath, 0o755)
		}
	}

	return nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && (!info.IsDir())
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

