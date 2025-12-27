//soppo:generated v1
package main

import "github.com/halcyonnouveau/soppo/runtime"
import "bufio"
import "fmt"
import "os"
import "path/filepath"
import "strings"
import slap "github.com/beanpuppy/slap/gen"
import "github.com/halcyonnouveau/sopmod/gen/internal/compat"
import "github.com/halcyonnouveau/sopmod/gen/internal/config"
import "github.com/halcyonnouveau/sopmod/gen/internal/install"
import "github.com/halcyonnouveau/sopmod/gen/internal/paths"
import "github.com/halcyonnouveau/sopmod/gen/internal/shim"

// Set at build time with -ldflags "-X main.version=v0.2.0"
var version = "dev"

// Install a Go or sop version
type InstallCmd struct {
	Tool string
	Version string
	Verbose bool
}

func (cmd InstallCmd) Run() error {
	switch cmd.Tool {
	case "go":
		_, _err0 := install.InstallGo(cmd.Version, cmd.Verbose)
		if _err0 != nil {
			return _err0
		}
	case "sop":
		resolved, _err1 := install.InstallSop(cmd.Version, cmd.Verbose)
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
		return fmt.Errorf("unknown tool '%s'. Use 'go' or 'sop'", cmd.Tool)
	}

	return nil
}

// List installed versions
type ListCmd struct {
	Tool string
}

func (cmd ListCmd) Run() error {
	cfg := config.Load()

	if cmd.Tool == "" {
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
	} else {
		switch cmd.Tool {
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
			return fmt.Errorf("unknown tool '%s'. Use 'go' or 'sop'", cmd.Tool)
		}
	}
	return nil
}

// Set the default sop version
type DefaultCmd struct {
	Version string
}

func (cmd DefaultCmd) Run() error {
	// Resolve version first
	resolved, _err0 := install.ResolveSopVersion(cmd.Version)
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

// Remove an installed version
type RemoveCmd struct {
	Tool string
	Version string
}

func (cmd RemoveCmd) Run() error {
	switch cmd.Tool {
	case "go":
		resolved, _err0 := install.ResolveGoVersion(cmd.Version)
		if _err0 != nil {
			return _err0
		}
		return install.RemoveGo(resolved)
	case "sop":
		resolved, _err1 := install.ResolveSopVersion(cmd.Version)
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
		return fmt.Errorf("unknown tool '%s'. Use 'go' or 'sop'", cmd.Tool)
	}
}

// Update to the latest version
type UpdateCmd struct {
	Tool string
}

func (cmd UpdateCmd) Run() error {
	if cmd.Tool == "" {
		_err0 := updateGo()
		if _err0 != nil {
			return _err0
		}
		return updateSop()
	}

	switch cmd.Tool {
	case "go":
		return updateGo()
	case "sop":
		return updateSop()
	default:
		return fmt.Errorf("unknown tool '%s'. Use 'go' or 'sop'", cmd.Tool)
	}
}

// All subcommands
/*soppo:enum
Cmd {
    Install InstallCmd
    List ListCmd
    Default DefaultCmd
    Remove RemoveCmd
    Update UpdateCmd
}
*/
type Cmd interface {
	isCmd()
}

type Cmd_Install struct {
	Value InstallCmd
}
func (Cmd_Install) isCmd() {}

type Cmd_List struct {
	Value ListCmd
}
func (Cmd_List) isCmd() {}

type Cmd_Default struct {
	Value DefaultCmd
}
func (Cmd_Default) isCmd() {}

type Cmd_Remove struct {
	Value RemoveCmd
}
func (Cmd_Remove) isCmd() {}

type Cmd_Update struct {
	Value UpdateCmd
}
func (Cmd_Update) isCmd() {}

func CmdInstall(value InstallCmd) Cmd {
	return Cmd_Install{Value: value}
}
func CmdList(value ListCmd) Cmd {
	return Cmd_List{Value: value}
}
func CmdDefault(value DefaultCmd) Cmd {
	return Cmd_Default{Value: value}
}
func CmdRemove(value RemoveCmd) Cmd {
	return Cmd_Remove{Value: value}
}
func CmdUpdate(value UpdateCmd) Cmd {
	return Cmd_Update{Value: value}
}

func main() {
	// Check if running as shim (invoked as "sop" not "sopmod")
	arg0 := os.Args[0]
	base := filepath.Base(arg0)
	isShim := strings.HasSuffix(base, "sop") && (!strings.HasSuffix(base, "sopmod"))

	if isShim {
		_err0 := shim.Run()
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

	_err2 := slap.Run[Cmd]()
	if _err2 != nil {
		err := _err2
		fmt.Fprintf(os.Stderr, "\033[31;1merror:\033[0m %s\n", err)
		os.Exit(1)
	}
}

func setDefaultSop(version string) error {
	cfg := config.Load()
	cfg.DefaultSop = (&version)

	// Install shim
	err := shim.Install()
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

		_err2 := shim.Install()
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
			if best == "" || shim.CompareVersions(v, best) > 0 {
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

func init() {
	runtime.RegisterAttr("main.InstallCmd", "", slap.Command{Name: "install", About: "Install a Go or sop version"})
	runtime.RegisterAttr("main.InstallCmd", "Tool", slap.Arg{Position: 0, Help: "Tool to install (go or sop)"})
	runtime.RegisterAttr("main.InstallCmd", "Version", slap.Arg{Position: 1, Help: "Version to install (e.g. latest, 1.23.0)"})
	runtime.RegisterAttr("main.InstallCmd", "Verbose", slap.Flag{Short: "v", Long: "verbose", Help: "Verbose output"})
	runtime.RegisterAttr("main.ListCmd", "", slap.Command{Name: "list", About: "List installed versions"})
	runtime.RegisterAttr("main.ListCmd", "Tool", slap.Arg{Position: 0, Help: "Tool to list (go or sop, omit for both)", Optional: true})
	runtime.RegisterAttr("main.DefaultCmd", "", slap.Command{Name: "default", About: "Set the default sop version"})
	runtime.RegisterAttr("main.DefaultCmd", "Version", slap.Arg{Position: 0, Help: "Version to set as default"})
	runtime.RegisterAttr("main.RemoveCmd", "", slap.Command{Name: "remove", About: "Remove an installed version"})
	runtime.RegisterAttr("main.RemoveCmd", "Tool", slap.Arg{Position: 0, Help: "Tool to remove (go or sop)"})
	runtime.RegisterAttr("main.RemoveCmd", "Version", slap.Arg{Position: 1, Help: "Version to remove"})
	runtime.RegisterAttr("main.UpdateCmd", "", slap.Command{Name: "update", About: "Update to the latest version"})
	runtime.RegisterAttr("main.UpdateCmd", "Tool", slap.Arg{Position: 0, Help: "Tool to update (go or sop, omit for both)", Optional: true})
	runtime.RegisterAttr("main.Cmd", "", slap.Command{Name: "sopmod", About: "Soppo version manager", Version: version})
	runtime.RegisterAttr("main.Cmd", "", slap.Subcommands{})
	runtime.RegisterAttr("main.Cmd", "Install", runtime.EnumVariant{WrapperType: Cmd_Install{}})
	runtime.RegisterAttr("main.Cmd", "List", runtime.EnumVariant{WrapperType: Cmd_List{}})
	runtime.RegisterAttr("main.Cmd", "Default", runtime.EnumVariant{WrapperType: Cmd_Default{}})
	runtime.RegisterAttr("main.Cmd", "Remove", runtime.EnumVariant{WrapperType: Cmd_Remove{}})
	runtime.RegisterAttr("main.Cmd", "Update", runtime.EnumVariant{WrapperType: Cmd_Update{}})
}
