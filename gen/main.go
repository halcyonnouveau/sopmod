//soppo:generated
package main

import "bufio"
import "flag"
import "fmt"
import "os"
import "path/filepath"
import "strings"
import "syscall"
import "github.com/BurntSushi/toml"
import "github.com/halcyonnouveau/sopmod/gen/internal/compat"
import "github.com/halcyonnouveau/sopmod/gen/internal/config"
import "github.com/halcyonnouveau/sopmod/gen/internal/install"
import "github.com/halcyonnouveau/sopmod/gen/internal/paths"

// Set at build time with -ldflags "-X main.version=v0.2.0"
var version = "dev"

func main() {
    // Check if running as shim (invoked as "sop" not "sopmod")
    arg0 := os.Args[0]
    base := filepath.Base(arg0)
    isShim := strings.HasSuffix(base, "sop") && (!strings.HasSuffix(base, "sopmod"))

    if isShim {
        _err0 := runShim()
        if _err0 != nil {
            err := _err0
            fmt.Fprintf(os.Stderr, "\033[31;1merror:\033[0m %s\n", err)
            os.Exit(1)
        }
        return
    }

    // Ensure sopmod directories exist
    _err1 := paths.EnsureDirs()
    if _err1 != nil {
        err := _err1
        fmt.Fprintf(os.Stderr, "Failed to create sopmod directories: %s\n", err)
        os.Exit(1)
    }

    if len(os.Args) < 2 {
        printUsage()
        os.Exit(1)
    }

    cmd := os.Args[1]
    args := os.Args[2:]

    var err error
    switch cmd {
    case "install":
        err = cmdInstall(args)
    case "list":
        err = cmdList(args)
    case "default":
        err = cmdDefault(args)
    case "remove":
        err = cmdRemove(args)
    case "update":
        err = cmdUpdate(args)
    case "help", "-h", "--help":
        printUsage()
        return
    case "version", "-v", "--version":
        fmt.Println("sopmod " + version)
        return
    default:
        fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
        printUsage()
        os.Exit(1)
    }

    if err != nil {
        fmt.Fprintf(os.Stderr, "\033[31;1merror:\033[0m %s\n", err)
        os.Exit(1)
    }
}

func printUsage() {
    fmt.Println("SOPMOD II - Soppo version manager")
    fmt.Println()
    fmt.Println("Usage: sopmod <command> [options]")
    fmt.Println()
    fmt.Println("Commands:")
    fmt.Println("  install <tool> <version>  Install a Go or sop version")
    fmt.Println("  list [tool]               List installed versions")
    fmt.Println("  default <version>         Set the default sop version")
    fmt.Println("  remove <tool> <version>   Remove an installed version")
    fmt.Println("  update [tool]             Update to the latest version")
}

func cmdInstall(args []string) error {
    fs := flag.NewFlagSet("install", flag.ExitOnError)
    verbose := fs.Bool("v", false, "Verbose output")
    fs.Parse(args)

    remaining := fs.Args()
    if len(remaining) < 2 {
        return fmt.Errorf("usage: sopmod install <tool> <version>")
    }

    tool := remaining[0]
    version := remaining[1]

    switch tool {
    case "go":
        _, _err0 := install.InstallGo(version, (*verbose))
        if _err0 != nil {
            return _err0
        }
    case "sop":
        resolved, _err1 := install.InstallSop(version, (*verbose))
        if _err1 != nil {
            return _err1
        }
        // Set as default if no default exists
        cfg := config.Load()
        if cfg.DefaultSop == nil {
            fmt.Printf("\033[36m→\033[0m Setting sop \033[1m%s\033[0m as default (first install)\n", resolved)
            return setDefaultSop(resolved)
        }
    default:
        return fmt.Errorf("unknown tool '%s'. Use 'go' or 'sop'", tool)
    }

    return nil
}

func cmdList(args []string) error {
    cfg := config.Load()

    var tool *string
    if len(args) > 0 {
        tool = (&args[0])
    }

    switch tool {
    case nil:
        // List both
        goVersions := install.ListInstalledGo()
        sopVersions := install.ListInstalledSop()
        if len(goVersions) == 0 && len(sopVersions) == 0 {
            fmt.Println("\033[2mNo versions installed\033[0m")
            fmt.Println("  \033[36mhint:\033[0m run \033[1msopmod install go latest\033[0m to install go")
            fmt.Println("  \033[36mhint:\033[0m run \033[1msopmod install sop latest\033[0m to install sop")
        } else {
            if len(goVersions) > 0 {
                fmt.Println("\033[1mgo:\033[0m")
                for _, v := range goVersions {
                    fmt.Printf("  %s\n", v)
                }
            }
            if len(sopVersions) > 0 {
                fmt.Println("\033[1msop:\033[0m")
                for _, v := range sopVersions {
                    if cfg.DefaultSop != nil && (*cfg.DefaultSop) == v {
                        fmt.Printf("  \033[32m%s\033[0m \033[2m(default)\033[0m\n", v)
                    } else {
                        fmt.Printf("  %s\n", v)
                    }
                }
            }
        }
    default:
        switch (*tool) {
        case "go":
            versions := install.ListInstalledGo()
            if len(versions) == 0 {
                fmt.Println("\033[2mNo go versions installed\033[0m")
            } else {
                fmt.Println("\033[1mInstalled go versions:\033[0m")
                for _, v := range versions {
                    fmt.Printf("  %s\n", v)
                }
            }
        case "sop":
            versions := install.ListInstalledSop()
            if len(versions) == 0 {
                fmt.Println("\033[2mNo sop versions installed\033[0m")
            } else {
                fmt.Println("\033[1mInstalled sop versions:\033[0m")
                for _, v := range versions {
                    if cfg.DefaultSop != nil && (*cfg.DefaultSop) == v {
                        fmt.Printf("  \033[32m%s\033[0m \033[2m(default)\033[0m\n", v)
                    } else {
                        fmt.Printf("  %s\n", v)
                    }
                }
            }
        default:
            return fmt.Errorf("unknown tool '%s'. Use 'go' or 'sop'", (*tool))
        }
    }
    return nil
}

func cmdDefault(args []string) error {
    if len(args) < 1 {
        return fmt.Errorf("usage: sopmod default <version>")
    }

    version := args[0]

    // Resolve version first
    resolved, _err0 := install.ResolveSopVersion(version)
    if _err0 != nil {
        return _err0
    }

    // Check if installed
    versions := install.ListInstalledSop()
    found := false
    for _, v := range versions {
        if v == resolved {
            found = true
            break
        }
    }

    if (!found) {
        shouldInstall, _err1 := promptInstall("sop", resolved)
        if _err1 != nil {
            return _err1
        }

        if shouldInstall {
            _, _err2 := install.InstallSop(resolved, false)
            if _err2 != nil {
                return _err2
            }
        } else {
            return nil
        }
    }

    return setDefaultSop(resolved)
}

func setDefaultSop(version string) error {
    cfg := config.Load()
    cfg.DefaultSop = (&version)

    // Install shim
    err := installShim()
    if err != nil {
        return err
    }

    fmt.Printf("\033[32m✓\033[0m Default sop version set to \033[1m%s\033[0m\n", version)

    // Auto-set compatible Go version
    goVersion, _err0 := findOrInstallCompatibleGo(version)
    if _err0 != nil {
        return _err0
    }

    if goVersion != "" {
        cfg.DefaultGo = (&goVersion)
        fmt.Printf("\033[32m✓\033[0m Default go version set to \033[1m%s\033[0m (compatible with sop %s)\n", goVersion, version)
    }

    _err1 := cfg.Save()
    if _err1 != nil {
        return _err1
    }

    printPathHint()
    return nil
}

func promptInstall(tool string, version string) (bool, error) {
    fmt.Printf("\033[1m%s\033[0m \033[1m%s\033[0m is not installed. Install it? [Y/n] ", tool, version)

    reader := bufio.NewReader(os.Stdin)
    input, err := reader.ReadString('\n')
    if err != nil {
        return false, err
    }

    input = strings.ToLower(strings.TrimSpace(input))
    return input == "" || input == "y" || input == "yes", nil
}

func printPathHint() {
    binDir := paths.BinDir()
    pathVar := os.Getenv("PATH")
    if strings.Contains(pathVar, binDir) {
        return
    }

    fmt.Println()
    fmt.Printf("\033[36mhint:\033[0m Add \033[1m~/.sopmod/bin\033[0m to your PATH:\n")
    fmt.Println()
    fmt.Println("  \033[2mexport PATH=\"$HOME/.sopmod/bin:$PATH\"\033[0m")
}

func cmdRemove(args []string) error {
    if len(args) < 2 {
        return fmt.Errorf("usage: sopmod remove <tool> <version>")
    }

    tool := args[0]
    version := args[1]

    switch tool {
    case "go":
        resolved, _err0 := install.ResolveGoVersion(version)
        if _err0 != nil {
            return _err0
        }
        return install.RemoveGo(resolved)
    case "sop":
        resolved, _err1 := install.ResolveSopVersion(version)
        if _err1 != nil {
            return _err1
        }
        _err2 := install.RemoveSop(resolved)
        if _err2 != nil {
            return _err2
        }
        // Clear default if it was this version
        cfg := config.Load()
        if cfg.DefaultSop != nil && (*cfg.DefaultSop) == resolved {
            cfg.DefaultSop = nil
            cfg.Save()
        }
        return nil
    default:
        return fmt.Errorf("unknown tool '%s'. Use 'go' or 'sop'", tool)
    }
}

func cmdUpdate(args []string) error {
    var tool *string
    if len(args) > 0 {
        tool = (&args[0])
    }

    switch tool {
    case nil:
        _err0 := updateGo()
        if _err0 != nil {
            return _err0
        }
        return updateSop()
    default:
        switch (*tool) {
        case "go":
            return updateGo()
        case "sop":
            return updateSop()
        default:
            return fmt.Errorf("unknown tool '%s'. Use 'go' or 'sop'", (*tool))
        }
    }
}

func updateGo() error {
    latest, _err0 := install.ResolveGoVersion("latest")
    if _err0 != nil {
        return _err0
    }

    installed := install.ListInstalledGo()
    for _, v := range installed {
        if v == latest {
            fmt.Printf("\033[32m✓\033[0m go \033[1m%s\033[0m is already the latest version\n", latest)
            return nil
        }
    }

    _, err := install.InstallGo(latest, false)
    return err
}

func updateSop() error {
    cfg := config.Load()
    oldDefault := cfg.DefaultSop

    latest, _err0 := install.ResolveSopVersion("latest")
    if _err0 != nil {
        return _err0
    }

    installed := install.ListInstalledSop()
    alreadyInstalled := false
    for _, v := range installed {
        if v == latest {
            alreadyInstalled = true
            break
        }
    }

    if alreadyInstalled {
        fmt.Printf("\033[32m✓\033[0m sop \033[1m%s\033[0m is already the latest version\n", latest)
    } else {
        _, _err1 := install.InstallSop(latest, false)
        if _err1 != nil {
            return _err1
        }
    }

    // Update default if needed
    shouldUpdateDefault := oldDefault == nil
    if oldDefault != nil {
        for _, v := range installed {
            if v == (*oldDefault) {
                shouldUpdateDefault = true
                break
            }
        }
    }

    if shouldUpdateDefault && (cfg.DefaultSop == nil || (*cfg.DefaultSop) != latest) {
        cfg.DefaultSop = (&latest)

        _err2 := installShim()
        if _err2 != nil {
            return _err2
        }

        fmt.Printf("\033[32m✓\033[0m Default sop version updated to \033[1m%s\033[0m\n", latest)

        goVersion, _err3 := findOrInstallCompatibleGo(latest)
        if _err3 != nil {
            return _err3
        }

        if goVersion != "" {
            cfg.DefaultGo = (&goVersion)
            fmt.Printf("\033[32m✓\033[0m Default go version set to \033[1m%s\033[0m (compatible with sop %s)\n", goVersion, latest)
        }

        cfg.Save()
    }

    return nil
}

func findOrInstallCompatibleGo(sopVersion string) (string, error) {
    compatInfo := compat.GoCompatFor(sopVersion)
    if compatInfo == nil {
        return "", nil
    }

    installedGo := install.ListInstalledGo()

    // Find best compatible installed version
    var best string
    for _, v := range installedGo {
        if compat.IsGoCompatible(v, sopVersion) {
            if best == "" || compareVersions(v, best) > 0 {
                best = v
            }
        }
    }

    if best != "" {
        return best, nil
    }

    // No compatible Go installed, install latest
    fmt.Printf("\033[36m→\033[0m Installing go (sop %s requires \033[1m%s+\033[0m)...\n", sopVersion, compatInfo.Min)
    return install.InstallGo("latest", false)
}

func compareVersions(a string, b string) int {
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

func installShim() error {
    currentExe, _err0 := os.Executable()
    if _err0 != nil {
        return _err0
    }
    shimPath := paths.SopShim()

    // Remove existing shim
    os.Remove(shimPath)

    // Copy current binary to shim location
    _err1 := copyFile(currentExe, shimPath)
    if _err1 != nil {
        return _err1
    }

    // Make executable
    return os.Chmod(shimPath, 0o755)
}

func runShim() error {
    version, _err0 := findSopVersion()
    if _err0 != nil {
        return _err0
    }
    sopBinary := paths.SopBinary(version)

    if (!fileExists(sopBinary)) {
        return fmt.Errorf("sop %s is not installed. Run `sopmod install sop %s`", version, version)
    }

    // Exec sop with all original args
    args := append([]string{sopBinary}, os.Args[1:]...)
    return syscall.Exec(sopBinary, args, os.Environ())
}

func findSopVersion() (string, error) {
    // Check sop.mod in current dir and parents
    version, err := findSopModVersion()
    if err == nil && version != "" {
        return version, nil
    }

    // Fall back to default
    cfg := config.Load()
    if cfg.DefaultSop != nil {
        return (*cfg.DefaultSop), nil
    }

    return "", fmt.Errorf("no sop version configured. Run `sopmod install sop latest`")
}

func findSopModVersion() (string, error) {
    type SopMod struct {
        Sop *string `toml:"sop"`
    }

    current, _err0 := os.Getwd()
    if _err0 != nil {
        return "", _err0
    }

    for {
        sopModPath := filepath.Join(current, "sop.mod")
        if fileExists(sopModPath) {
            var modCfg SopMod
            _, err := toml.DecodeFile(sopModPath, (&modCfg))
            if err != nil {
                return "", err
            }
            if modCfg.Sop != nil {
                return (*modCfg.Sop), nil
            }
            return "", nil
        }

        parent := filepath.Dir(current)
        if parent == current {
            break
        }
        current = parent
    }

    return "", nil
}

func copyFile(src string, dst string) error {
    data, _err0 := os.ReadFile(src)
    if _err0 != nil {
        return _err0
    }
    return os.WriteFile(dst, data, 0o755)
}

func fileExists(path string) bool {
    info, err := os.Stat(path)
    return err == nil && (!info.IsDir())
}

