use serde::{Deserialize, Serialize};
use std::path::Path;

use crate::paths;

/// Global sopmod configuration stored in ~/.sopmod/config.toml
#[derive(Debug, Default, Serialize, Deserialize)]
pub struct Config {
    /// Default sop version to use
    pub default_sop: Option<String>,
    /// Default Go version (automatically set based on sop compatibility)
    pub default_go: Option<String>,
}

impl Config {
    /// Load config from ~/.sopmod/config.toml, or return default if not found
    pub fn load() -> Self {
        let path = paths::config_path();
        Self::load_from(&path).unwrap_or_default()
    }

    /// Load config from a specific path
    pub fn load_from(path: &Path) -> Option<Self> {
        let content = std::fs::read_to_string(path).ok()?;
        toml::from_str(&content).ok()
    }

    /// Save config to ~/.sopmod/config.toml
    pub fn save(&self) -> std::io::Result<()> {
        let path = paths::config_path();
        self.save_to(&path)
    }

    /// Save config to a specific path
    pub fn save_to(&self, path: &Path) -> std::io::Result<()> {
        let content =
            toml::to_string_pretty(self).map_err(|e| std::io::Error::other(e.to_string()))?;
        if let Some(parent) = path.parent() {
            std::fs::create_dir_all(parent)?;
        }
        std::fs::write(path, content)
    }
}

/// Project-specific version requirements from sop.mod
#[derive(Debug, Default, Deserialize)]
pub struct ProjectConfig {
    /// Required Go version
    pub go: Option<String>,
    /// Required sop version
    pub sop: Option<String>,
}

impl ProjectConfig {
    /// Load project config from sop.mod in the given directory
    pub fn load_from_dir(dir: &Path) -> Option<Self> {
        let path = dir.join("sop.mod");
        let content = std::fs::read_to_string(path).ok()?;
        toml::from_str(&content).ok()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::io::Write;

    #[test]
    fn test_config_roundtrip() {
        let dir = tempfile::tempdir().unwrap();
        let path = dir.path().join("config.toml");

        let config = Config {
            default_sop: Some("0.2.0".to_string()),
            default_go: Some("1.22.0".to_string()),
        };

        config.save_to(&path).unwrap();
        let loaded = Config::load_from(&path).unwrap();

        assert_eq!(loaded.default_sop, Some("0.2.0".to_string()));
        assert_eq!(loaded.default_go, Some("1.22.0".to_string()));
    }

    #[test]
    fn test_project_config() {
        let dir = tempfile::tempdir().unwrap();
        let path = dir.path().join("sop.mod");

        let mut file = std::fs::File::create(&path).unwrap();
        writeln!(file, "package = \"test\"").unwrap();
        writeln!(file, "go = \"1.22\"").unwrap();
        writeln!(file, "sop = \"0.4.0\"").unwrap();

        let config = ProjectConfig::load_from_dir(dir.path()).unwrap();
        assert_eq!(config.go, Some("1.22".to_string()));
        assert_eq!(config.sop, Some("0.4.0".to_string()));
    }
}
