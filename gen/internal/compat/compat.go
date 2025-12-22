//soppo:generated v1
package compat

import "fmt"
import "strconv"
import "strings"

// GoCompat holds Go version compatibility requirements for a sop version.
type GoCompat struct {
	Min string
	Max *string //soppo:nilable
}

// Minimum required Go version (inclusive)
// Maximum supported Go version (inclusive), nil means no upper bound
// GoCompatFor returns the Go version compatibility for a given sop version.
// Returns nil if the version is unknown or doesn't have specific requirements.
//
// ```sop
// import "fmt"
// compat := GoCompatFor("0.5.0")
// fmt.Println(compat.Min)
// // Output:
// // 1.21
// ```
//soppo:nilable : 0
func GoCompatFor(sopVersion string) *GoCompat {
	version := strings.TrimPrefix(sopVersion, "v")

	major, minor, _, _err0 := parseVersion(version)
	if _err0 != nil {
		return nil
	}

	// Add new Go requirements here (check newest first)
	// if major > 0 || (major == 0 && minor >= 6) {
	//     return &GoCompat{Min: "1.25", Max: nil}
	// }
	if major > 0 || (major == 0 && minor >= 0) {
		return (&GoCompat{Min: "1.21", Max: nil})
	}

	return (&GoCompat{Min: "1.21", Max: nil})
}

// IsGoCompatible checks if a Go version satisfies the requirements for a sop version.
//
// ```sop
// import "fmt"
// fmt.Println(IsGoCompatible("1.22.0", "0.5.0"))
// fmt.Println(IsGoCompatible("1.20.0", "0.5.0"))
// // Output:
// // true
// // false
// ```
func IsGoCompatible(goVersion string, sopVersion string) bool {
	compat := GoCompatFor(sopVersion)
	if compat == nil {
		// Unknown sop version, assume compatible
		return true
	}

	goMajor, goMinor, goPatch, _err0 := parseVersion(goVersion)
	if _err0 != nil {
		return false
	}

	minMajor, minMinor, minPatch, _err1 := parseVersion(compat.Min)
	if _err1 != nil {
		return false
	}

	// Check minimum
	if (!versionAtLeast(goMajor, goMinor, goPatch, minMajor, minMinor, minPatch)) {
		return false
	}

	// Check maximum if set
	if compat.Max != nil {
		maxMajor, maxMinor, maxPatch, _err2 := parseVersion((*compat.Max))
		if _err2 != nil {
			return false
		}

		if (!versionAtLeast(maxMajor, maxMinor, maxPatch, goMajor, goMinor, goPatch)) {
			return false
		}
	}

	return true
}

// CompatMessage returns a human-readable compatibility message for a sop version.
//
// ```sop
// import "fmt"
//
// fmt.Println(CompatMessage("0.5.0"))
// // Output:
// // sop 0.5.0 requires go 1.21 or later
// ```
func CompatMessage(sopVersion string) string {
	compat := GoCompatFor(sopVersion)
	if compat == nil {
		return fmt.Sprintf("sop %s has unknown go requirements", sopVersion)
	}
	if compat.Max != nil {
		return fmt.Sprintf("sop %s requires go %s to %s", sopVersion, compat.Min, (*compat.Max))
	}
	return fmt.Sprintf("sop %s requires go %s or later", sopVersion, compat.Min)
}

func parseVersion(version string) (major int, minor int, patch int, err error) {
	parts := strings.Split(version, ".")

	if len(parts) >= 1 {
		var _err0 error
		major, _err0 = strconv.Atoi(parts[0])
		if _err0 != nil {
			return 0, 0, 0, _err0
		}
	}
	if len(parts) >= 2 {
		var _err1 error
		minor, _err1 = strconv.Atoi(parts[1])
		if _err1 != nil {
			return 0, 0, 0, _err1
		}
	}
	if len(parts) >= 3 {
		var _err2 error
		patch, _err2 = strconv.Atoi(parts[2])
		if _err2 != nil {
			return 0, 0, 0, _err2
		}
	}

	return major, minor, patch, nil
}

func versionAtLeast(aMajor int, aMinor int, aPatch int, bMajor int, bMinor int, bPatch int) bool {
	if aMajor != bMajor {
		return aMajor > bMajor
	}
	if aMinor != bMinor {
		return aMinor > bMinor
	}
	return aPatch >= bPatch
}

