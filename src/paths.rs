use std::path::PathBuf;

/// Returns the sopmod root directory (~/.sopmod)
pub fn sopmod_dir() -> PathBuf {
    dirs::home_dir()
        .expect("Could not determine home directory")
        .join(".sopmod")
}

/// Returns the Go installations directory (~/.sopmod/go)
pub fn go_root() -> PathBuf {
    sopmod_dir().join("go")
}

/// Returns a specific Go version directory (~/.sopmod/go/1.22.0)
pub fn go_dir(version: &str) -> PathBuf {
    go_root().join(version)
}

/// Returns the Go binary path for a specific version
pub fn go_binary(version: &str) -> PathBuf {
    let dir = go_dir(version);
    if cfg!(windows) {
        dir.join("go").join("bin").join("go.exe")
    } else {
        dir.join("go").join("bin").join("go")
    }
}

/// Returns the sop installations directory (~/.sopmod/sop)
pub fn sop_root() -> PathBuf {
    sopmod_dir().join("sop")
}

/// Returns a specific sop version directory (~/.sopmod/sop/0.2.0)
pub fn sop_dir(version: &str) -> PathBuf {
    sop_root().join(version)
}

/// Returns the sop binary path for a specific version
pub fn sop_binary(version: &str) -> PathBuf {
    let dir = sop_dir(version);
    if cfg!(windows) {
        dir.join("sop.exe")
    } else {
        dir.join("sop")
    }
}

/// Returns the config file path (~/.sopmod/config.toml)
pub fn config_path() -> PathBuf {
    sopmod_dir().join("config.toml")
}

/// Returns the bin directory for symlinks (~/.sopmod/bin)
pub fn bin_dir() -> PathBuf {
    sopmod_dir().join("bin")
}

/// Returns the symlinked sop binary path (~/.sopmod/bin/sop)
pub fn sop_symlink() -> PathBuf {
    let dir = bin_dir();
    if cfg!(windows) {
        dir.join("sop.exe")
    } else {
        dir.join("sop")
    }
}

/// Ensures the sopmod directory structure exists
pub fn ensure_dirs() -> std::io::Result<()> {
    std::fs::create_dir_all(go_root())?;
    std::fs::create_dir_all(sop_root())?;
    std::fs::create_dir_all(bin_dir())?;
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_paths_are_under_sopmod() {
        let sopmod = sopmod_dir();
        assert!(go_root().starts_with(&sopmod));
        assert!(sop_root().starts_with(&sopmod));
        assert!(config_path().starts_with(&sopmod));
    }

    #[test]
    fn test_version_dirs() {
        let go = go_dir("1.22.0");
        assert!(go.ends_with("1.22.0"));
        assert!(go.starts_with(go_root()));

        let sop = sop_dir("0.2.0");
        assert!(sop.ends_with("0.2.0"));
        assert!(sop.starts_with(sop_root()));
    }
}
