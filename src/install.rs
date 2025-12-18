use std::fs::{self, File};
use std::io::{self, Read, Write};
use std::path::Path;

use console::style;
use flate2::read::GzDecoder;
use indicatif::{ProgressBar, ProgressStyle};
use reqwest::blocking::Client;
use serde::Deserialize;
use tar::Archive;
use thiserror::Error;

use crate::compat;
use crate::paths;

#[derive(Error, Debug)]
pub enum InstallError {
    #[error("HTTP request failed: {0}")]
    Http(#[from] reqwest::Error),

    #[error("IO error: {0}")]
    Io(#[from] io::Error),

    #[error("Failed to extract archive: {0}")]
    Extract(String),

    #[error("Version not found: {0}")]
    VersionNotFound(String),

    #[error("Unsupported platform: {os}-{arch}")]
    UnsupportedPlatform { os: String, arch: String },

    #[error("Zip error: {0}")]
    Zip(#[from] zip::result::ZipError),

    #[error("Incompatible versions: {0}")]
    Incompatible(String),
}

/// Platform information for downloads
struct Platform {
    os: &'static str,
    arch: &'static str,
}

impl Platform {
    fn detect() -> Result<Self, InstallError> {
        let os = if cfg!(target_os = "linux") {
            "linux"
        } else if cfg!(target_os = "macos") {
            "darwin"
        } else if cfg!(target_os = "windows") {
            "windows"
        } else {
            return Err(InstallError::UnsupportedPlatform {
                os: std::env::consts::OS.to_string(),
                arch: std::env::consts::ARCH.to_string(),
            });
        };

        let arch = if cfg!(target_arch = "x86_64") {
            "amd64"
        } else if cfg!(target_arch = "aarch64") {
            "arm64"
        } else {
            return Err(InstallError::UnsupportedPlatform {
                os: std::env::consts::OS.to_string(),
                arch: std::env::consts::ARCH.to_string(),
            });
        };

        Ok(Platform { os, arch })
    }

    fn go_archive_ext(&self) -> &'static str {
        if self.os == "windows" {
            "zip"
        } else {
            "tar.gz"
        }
    }
}

/// Response from go.dev/dl/?mode=json
#[derive(Debug, Deserialize)]
struct GoRelease {
    version: String,
    stable: bool,
}

/// Resolve "latest" to the actual latest stable Go version
pub fn resolve_latest_go() -> Result<String, InstallError> {
    let client = Client::new();
    let resp: Vec<GoRelease> = client.get("https://go.dev/dl/?mode=json").send()?.json()?;

    resp.into_iter()
        .find(|r| r.stable)
        .map(|r| {
            r.version
                .strip_prefix("go")
                .unwrap_or(&r.version)
                .to_string()
        })
        .ok_or_else(|| InstallError::VersionNotFound("latest".to_string()))
}

/// Resolve a Go version, handling "latest" and partial versions like "1.22"
pub fn resolve_go_version(version: &str) -> Result<String, InstallError> {
    if version == "latest" {
        return resolve_latest_go();
    }

    // Check if it's a partial version like "1.22" that needs resolution
    let parts: Vec<&str> = version.split('.').collect();
    if parts.len() == 2 {
        // Fetch available versions and find the latest patch
        let client = Client::new();
        let resp: Vec<GoRelease> = client.get("https://go.dev/dl/?mode=json").send()?.json()?;

        let prefix = format!("go{}", version);
        let latest = resp
            .into_iter()
            .filter(|r| r.stable && r.version.starts_with(&prefix))
            .map(|r| r.version)
            .next();

        if let Some(v) = latest {
            return Ok(v.strip_prefix("go").unwrap_or(v.as_str()).to_string());
        }
    }

    // Return as-is, assuming it's a full version
    Ok(version.to_string())
}

/// Install a specific Go version
pub fn install_go(version: &str, verbose: bool) -> Result<String, InstallError> {
    let version = resolve_go_version(version)?;
    let platform = Platform::detect()?;
    let dest = paths::go_dir(&version);

    if dest.exists() {
        if verbose {
            println!(
                "{} go {} is already installed",
                style("✓").green(),
                style(&version).bold()
            );
        }
        return Ok(version);
    }

    let ext = platform.go_archive_ext();
    let filename = format!("go{}.{}-{}.{}", version, platform.os, platform.arch, ext);
    let url = format!("https://go.dev/dl/{}", filename);

    if verbose {
        println!("Downloading go {} from {}", version, url);
    }

    // Download with progress
    let client = Client::new();
    let resp = client.get(&url).send()?;

    if !resp.status().is_success() {
        return Err(InstallError::VersionNotFound(format!(
            "go {} for {}-{}",
            version, platform.os, platform.arch
        )));
    }

    let total_size = resp.content_length().unwrap_or(0);
    let pb = create_progress_bar(total_size, &format!("Downloading go {}", version));

    // Download to temp file
    let temp_dir = tempfile::tempdir()?;
    let archive_path = temp_dir.path().join(&filename);
    let mut file = File::create(&archive_path)?;
    let mut downloaded: u64 = 0;

    let mut reader = resp;
    let mut buffer = [0u8; 8192];
    loop {
        let bytes_read = reader.read(&mut buffer)?;
        if bytes_read == 0 {
            break;
        }
        file.write_all(&buffer[..bytes_read])?;
        downloaded += bytes_read as u64;
        pb.set_position(downloaded);
    }
    pb.finish_with_message("Download complete");

    // Extract
    if verbose {
        println!("Extracting to {:?}", dest);
    }

    fs::create_dir_all(&dest)?;

    if ext == "zip" {
        extract_zip(&archive_path, &dest)?;
    } else {
        extract_tar_gz(&archive_path, &dest)?;
    }

    // Verify installation
    let go_bin = paths::go_binary(&version);
    if !go_bin.exists() {
        return Err(InstallError::Extract(format!(
            "go binary not found at {:?}",
            go_bin
        )));
    }

    println!(
        "{} go {} installed successfully",
        style("✓").green(),
        style(&version).bold()
    );
    Ok(version)
}

/// GitHub release response
#[derive(Debug, Deserialize)]
struct GitHubRelease {
    tag_name: String,
    assets: Vec<GitHubAsset>,
}

#[derive(Debug, Deserialize)]
struct GitHubAsset {
    name: String,
    browser_download_url: String,
}

/// Resolve "latest" to the actual latest sop version
pub fn resolve_latest_sop() -> Result<String, InstallError> {
    let client = Client::new();
    let resp: GitHubRelease = client
        .get("https://api.github.com/repos/halcyonnouveau/soppo/releases/latest")
        .header("User-Agent", "sopmod")
        .send()?
        .json()?;

    Ok(resp
        .tag_name
        .strip_prefix('v')
        .unwrap_or(&resp.tag_name)
        .to_string())
}

/// Resolve a sop version, handling "latest"
pub fn resolve_sop_version(version: &str) -> Result<String, InstallError> {
    if version == "latest" {
        resolve_latest_sop()
    } else {
        Ok(version.to_string())
    }
}

/// Install a specific sop version
pub fn install_sop(version: &str, verbose: bool) -> Result<String, InstallError> {
    let version = resolve_sop_version(version)?;
    let platform = Platform::detect()?;
    let dest = paths::sop_dir(&version);

    if dest.exists() {
        if verbose {
            println!(
                "{} sop {} is already installed",
                style("✓").green(),
                style(&version).bold()
            );
        }
        return Ok(version);
    }

    // Check Go compatibility
    if let Some(compat_info) = compat::go_compat(&version) {
        let installed_go = list_installed_go();
        let has_compatible = installed_go
            .iter()
            .any(|v| compat::is_go_compatible(v, &version).unwrap_or(false));

        if !has_compatible && !installed_go.is_empty() {
            eprintln!(
                "{} {}",
                style("warning:").yellow().bold(),
                compat::compat_message(&version)
            );
            eprintln!("  Installed go versions: {}", installed_go.join(", "));
            eprintln!(
                "  {} run {}",
                style("hint:").cyan(),
                style(format!("sopmod install go {}", compat_info.min)).bold()
            );
        }
    }

    // Map platform to Rust target triple (used in release asset names)
    let target_triple = match (platform.os, platform.arch) {
        ("linux", "amd64") => "x86_64-unknown-linux-gnu",
        ("linux", "arm64") => "aarch64-unknown-linux-gnu",
        ("darwin", "amd64") => "x86_64-apple-darwin",
        ("darwin", "arm64") => "aarch64-apple-darwin",
        ("windows", "amd64") => "x86_64-pc-windows-msvc",
        ("windows", "arm64") => "aarch64-pc-windows-msvc",
        _ => {
            return Err(InstallError::UnsupportedPlatform {
                os: platform.os.to_string(),
                arch: platform.arch.to_string(),
            });
        }
    };

    // Fetch release info from GitHub
    let client = Client::new();
    let tag = format!("v{}", version);
    let release_url = format!(
        "https://api.github.com/repos/halcyonnouveau/soppo/releases/tags/{}",
        tag
    );

    let resp = client
        .get(&release_url)
        .header("User-Agent", "sopmod")
        .send()?;

    if !resp.status().is_success() {
        return Err(InstallError::VersionNotFound(format!("sop {}", version)));
    }

    let release: GitHubRelease = resp.json()?;

    // Find the right asset
    let asset_pattern = format!("sop-{}", target_triple);
    let asset = release
        .assets
        .iter()
        .find(|a| a.name.starts_with(&asset_pattern))
        .ok_or_else(|| {
            InstallError::VersionNotFound(format!("sop {} for {}", version, target_triple))
        })?;

    if verbose {
        println!(
            "Downloading sop {} from {}",
            version, asset.browser_download_url
        );
    }

    // Download with progress
    let resp = client
        .get(&asset.browser_download_url)
        .header("User-Agent", "sopmod")
        .send()?;

    if !resp.status().is_success() {
        return Err(InstallError::Http(resp.error_for_status().unwrap_err()));
    }

    let total_size = resp.content_length().unwrap_or(0);
    let pb = create_progress_bar(total_size, &format!("Downloading sop {}", version));

    // Download to temp file
    let temp_dir = tempfile::tempdir()?;
    let archive_path = temp_dir.path().join(&asset.name);
    let mut file = File::create(&archive_path)?;
    let mut downloaded: u64 = 0;

    let mut reader = resp;
    let mut buffer = [0u8; 8192];
    loop {
        let bytes_read = reader.read(&mut buffer)?;
        if bytes_read == 0 {
            break;
        }
        file.write_all(&buffer[..bytes_read])?;
        downloaded += bytes_read as u64;
        pb.set_position(downloaded);
    }
    pb.finish_with_message("Download complete");

    // Extract
    if verbose {
        println!("Extracting to {:?}", dest);
    }

    fs::create_dir_all(&dest)?;

    if asset.name.ends_with(".zip") {
        extract_zip(&archive_path, &dest)?;
    } else if asset.name.ends_with(".tar.gz") || asset.name.ends_with(".tgz") {
        extract_tar_gz(&archive_path, &dest)?;
    } else {
        // Assume it's a raw binary
        let binary_name = if cfg!(windows) { "sop.exe" } else { "sop" };
        let dest_path = dest.join(binary_name);
        fs::copy(&archive_path, &dest_path)?;
        #[cfg(unix)]
        {
            use std::os::unix::fs::PermissionsExt;
            fs::set_permissions(&dest_path, fs::Permissions::from_mode(0o755))?;
        }
    }

    // Verify installation
    let sop_bin = paths::sop_binary(&version);
    if !sop_bin.exists() {
        return Err(InstallError::Extract(format!(
            "sop binary not found at {:?}",
            sop_bin
        )));
    }

    println!(
        "{} sop {} installed successfully",
        style("✓").green(),
        style(&version).bold()
    );
    Ok(version)
}

/// List installed Go versions
pub fn list_installed_go() -> Vec<String> {
    let go_root = paths::go_root();
    if !go_root.exists() {
        return vec![];
    }

    fs::read_dir(go_root)
        .ok()
        .map(|entries| {
            entries
                .filter_map(|e| e.ok())
                .filter(|e| e.path().is_dir())
                .filter_map(|e| e.file_name().into_string().ok())
                .collect()
        })
        .unwrap_or_default()
}

/// List installed sop versions
pub fn list_installed_sop() -> Vec<String> {
    let sop_root = paths::sop_root();
    if !sop_root.exists() {
        return vec![];
    }

    fs::read_dir(sop_root)
        .ok()
        .map(|entries| {
            entries
                .filter_map(|e| e.ok())
                .filter(|e| e.path().is_dir())
                .filter_map(|e| e.file_name().into_string().ok())
                .collect()
        })
        .unwrap_or_default()
}

/// Remove an installed Go version
pub fn remove_go(version: &str) -> Result<(), InstallError> {
    let dir = paths::go_dir(version);
    if !dir.exists() {
        return Err(InstallError::VersionNotFound(format!("go {}", version)));
    }
    fs::remove_dir_all(dir)?;
    println!(
        "{} Removed go {}",
        style("✓").green(),
        style(version).bold()
    );
    Ok(())
}

/// Remove an installed sop version
pub fn remove_sop(version: &str) -> Result<(), InstallError> {
    let dir = paths::sop_dir(version);
    if !dir.exists() {
        return Err(InstallError::VersionNotFound(format!("sop {}", version)));
    }
    fs::remove_dir_all(dir)?;
    println!(
        "{} Removed sop {}",
        style("✓").green(),
        style(version).bold()
    );
    Ok(())
}

fn create_progress_bar(total: u64, msg: &str) -> ProgressBar {
    let pb = ProgressBar::new(total);
    pb.set_style(
        ProgressStyle::default_bar()
            .template("{msg}\n{bar:40.cyan/blue} {bytes}/{total_bytes} ({eta})")
            .unwrap()
            .progress_chars("##-"),
    );
    pb.set_message(msg.to_string());
    pb
}

fn extract_tar_gz(archive_path: &Path, dest: &Path) -> Result<(), InstallError> {
    let file = File::open(archive_path)?;
    let decoder = GzDecoder::new(file);
    let mut archive = Archive::new(decoder);
    archive
        .unpack(dest)
        .map_err(|e| InstallError::Extract(e.to_string()))
}

fn extract_zip(archive_path: &Path, dest: &Path) -> Result<(), InstallError> {
    let file = File::open(archive_path)?;
    let mut archive = zip::ZipArchive::new(file)?;

    for i in 0..archive.len() {
        let mut file = archive.by_index(i)?;
        let outpath = dest.join(file.mangled_name());

        if file.name().ends_with('/') {
            fs::create_dir_all(&outpath)?;
        } else {
            if let Some(p) = outpath.parent()
                && !p.exists()
            {
                fs::create_dir_all(p)?;
            }
            let mut outfile = File::create(&outpath)?;
            io::copy(&mut file, &mut outfile)?;
        }

        #[cfg(unix)]
        {
            use std::os::unix::fs::PermissionsExt;
            if let Some(mode) = file.unix_mode() {
                fs::set_permissions(&outpath, fs::Permissions::from_mode(mode))?;
            }
        }
    }

    Ok(())
}
