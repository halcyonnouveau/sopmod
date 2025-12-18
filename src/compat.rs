use semver::Version;

/// Go version compatibility requirements for a sop version
#[derive(Debug, Clone)]
pub struct GoCompat {
    /// Minimum required Go version (inclusive)
    pub min: &'static str,
    /// Maximum supported Go version (inclusive), None means no upper bound
    pub max: Option<&'static str>,
}

/// Returns the Go version compatibility for a given sop version
pub fn go_compat(sop_version: &str) -> Option<GoCompat> {
    let version = sop_version.strip_prefix('v').unwrap_or(sop_version);

    // Parse sop version
    let sop_ver = match Version::parse(&normalise_version(version)) {
        Ok(v) => v,
        Err(_) => return None,
    };

    // Add new Go requirements here (check newest first)
    // if sop_ver >= Version::new(0, 6, 0) {
    //     return Some(GoCompat { min: "1.25", max: None });
    // }

    if sop_ver >= Version::new(0, 1, 0) {
        return Some(GoCompat {
            min: "1.21",
            max: None,
        });
    }

    None
}

fn normalise_version(version: &str) -> String {
    let parts: Vec<&str> = version.split('.').collect();
    match parts.len() {
        1 => format!("{}.0.0", version),
        2 => format!("{}.0", version),
        _ => version.to_string(),
    }
}

/// Check if a Go version satisfies the requirements for a sop version
pub fn is_go_compatible(go_version: &str, sop_version: &str) -> Result<bool, String> {
    let Some(compat) = go_compat(sop_version) else {
        // Unknown sop version, assume compatible
        return Ok(true);
    };

    let go_ver = parse_go_version(go_version)?;
    let min_ver = parse_go_version(compat.min)?;

    if go_ver < min_ver {
        return Ok(false);
    }

    if let Some(max) = compat.max {
        let max_ver = parse_go_version(max)?;
        if go_ver > max_ver {
            return Ok(false);
        }
    }

    Ok(true)
}

/// Parse a Go version string (e.g., "1.22.0" or "1.22") into a semver Version
fn parse_go_version(version: &str) -> Result<Version, String> {
    // Go versions can be "1.22" or "1.22.0", normalise to semver
    let parts: Vec<&str> = version.split('.').collect();
    let normalised = match parts.len() {
        2 => format!("{}.0", version),
        3 => version.to_string(),
        _ => return Err(format!("Invalid go version format: {}", version)),
    };

    Version::parse(&normalised).map_err(|e| format!("Failed to parse version '{}': {}", version, e))
}

/// Get a human-readable compatibility message for a sop version
pub fn compat_message(sop_version: &str) -> String {
    match go_compat(sop_version) {
        Some(compat) => match compat.max {
            Some(max) => format!("sop {} requires go {} to {}", sop_version, compat.min, max),
            None => format!("sop {} requires go {} or later", sop_version, compat.min),
        },
        None => format!("sop {} has unknown go requirements", sop_version),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_go_compat() {
        assert!(go_compat("0.4.1").is_some());
        assert!(go_compat("v0.4.1").is_some());
        assert!(go_compat("0.1.0").is_some());
        assert!(go_compat("99.0.0").is_some()); // All versions require Go 1.21+
    }

    #[test]
    fn test_is_go_compatible() {
        // Go 1.22 should be compatible with sop 0.4.1 (requires 1.21+)
        assert!(is_go_compatible("1.22.0", "0.4.1").unwrap());
        assert!(is_go_compatible("1.22", "0.4.1").unwrap());
        assert!(is_go_compatible("1.21.0", "0.4.1").unwrap());

        // Go 1.20 should not be compatible
        assert!(!is_go_compatible("1.20.0", "0.4.1").unwrap());
        assert!(!is_go_compatible("1.19", "0.4.1").unwrap());
    }

    #[test]
    fn test_parse_go_version() {
        assert!(parse_go_version("1.22.0").is_ok());
        assert!(parse_go_version("1.22").is_ok());
        assert!(parse_go_version("1.21.5").is_ok());
    }
}
