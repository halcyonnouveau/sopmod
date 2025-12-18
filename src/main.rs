use clap::{Parser, Subcommand};
use console::style;

use sopmod::compat;
use sopmod::config::Config;
use sopmod::install;
use sopmod::paths;

#[derive(Parser)]
#[command(name = "sopmod")]
#[command(about = "Version manager for go and sop", long_about = None)]
struct Cli {
    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    /// Install a Go or sop version
    Install {
        /// The tool to install (go or sop)
        tool: String,
        /// The version to install (e.g., 1.22, 0.4.1, latest)
        version: String,
        /// Verbose output
        #[arg(short, long)]
        verbose: bool,
    },

    /// List installed versions
    List {
        /// The tool to list (go or sop), or omit for both
        tool: Option<String>,
    },

    /// Set the default version for a tool
    Default {
        /// The tool (go or sop)
        tool: String,
        /// The version to set as default
        version: String,
    },

    /// Show the path to the active version of a tool
    Which {
        /// The tool (go or sop)
        tool: String,
    },

    /// Remove an installed version
    Remove {
        /// The tool (go or sop)
        tool: String,
        /// The version to remove
        version: String,
    },

    /// Update to the latest version
    Update {
        /// The tool to update (go or sop), or omit for both
        tool: Option<String>,
    },
}

fn main() {
    // Ensure sopmod directories exist
    if let Err(e) = paths::ensure_dirs() {
        eprintln!("Failed to create sopmod directories: {}", e);
        std::process::exit(1);
    }

    let cli = Cli::parse();

    let result = match cli.command {
        Commands::Install {
            tool,
            version,
            verbose,
        } => cmd_install(&tool, &version, verbose),
        Commands::List { tool } => cmd_list(tool.as_deref()),
        Commands::Default { tool, version } => cmd_default(&tool, &version),
        Commands::Which { tool } => cmd_which(&tool),
        Commands::Remove { tool, version } => cmd_remove(&tool, &version),
        Commands::Update { tool } => cmd_update(tool.as_deref()),
    };

    if let Err(e) = result {
        eprintln!("{} {}", style("error:").red().bold(), e);
        std::process::exit(1);
    }
}

fn cmd_install(tool: &str, version: &str, verbose: bool) -> Result<(), Box<dyn std::error::Error>> {
    match tool {
        "go" => {
            install::install_go(version, verbose)?;
        }
        "sop" => {
            install::install_sop(version, verbose)?;
        }
        _ => {
            return Err(format!("Unknown tool '{}'. Use 'go' or 'sop'.", tool).into());
        }
    }
    Ok(())
}

fn cmd_list(tool: Option<&str>) -> Result<(), Box<dyn std::error::Error>> {
    let config = Config::load();

    match tool {
        Some("go") => {
            let versions = install::list_installed_go();
            if versions.is_empty() {
                println!("{}", style("No go versions installed").dim());
            } else {
                println!("{}", style("Installed go versions:").bold());
                for v in versions {
                    println!("  {}", v);
                }
            }
        }
        Some("sop") => {
            let versions = install::list_installed_sop();
            if versions.is_empty() {
                println!("{}", style("No sop versions installed").dim());
            } else {
                println!("{}", style("Installed sop versions:").bold());
                for v in versions {
                    if config.default_sop.as_deref() == Some(&v) {
                        println!("  {} {}", style(&v).green(), style("(default)").dim());
                    } else {
                        println!("  {}", v);
                    }
                }
            }
        }
        None => {
            // List both
            let go_versions = install::list_installed_go();
            let sop_versions = install::list_installed_sop();

            if go_versions.is_empty() && sop_versions.is_empty() {
                println!("{}", style("No versions installed").dim());
                println!(
                    "  {} run {} to install go",
                    style("hint:").cyan(),
                    style("sopmod install go latest").bold()
                );
                println!(
                    "  {} run {} to install sop",
                    style("hint:").cyan(),
                    style("sopmod install sop latest").bold()
                );
            } else {
                if !go_versions.is_empty() {
                    println!("{}", style("go:").bold());
                    for v in go_versions {
                        println!("  {}", v);
                    }
                }
                if !sop_versions.is_empty() {
                    println!("{}", style("sop:").bold());
                    for v in sop_versions {
                        if config.default_sop.as_deref() == Some(&v) {
                            println!("  {} {}", style(&v).green(), style("(default)").dim());
                        } else {
                            println!("  {}", v);
                        }
                    }
                }
            }
        }
        Some(other) => {
            return Err(format!("Unknown tool '{}'. Use 'go' or 'sop'.", other).into());
        }
    }
    Ok(())
}

fn cmd_default(tool: &str, version: &str) -> Result<(), Box<dyn std::error::Error>> {
    let mut config = Config::load();

    match tool {
        "go" => {
            return Err("go versions are managed automatically by sop. Use `sopmod install go <version>` to install.".into());
        }
        "sop" => {
            // Resolve version first (handles "latest")
            let resolved = install::resolve_sop_version(version)?;

            // Check if installed, offer to install if not
            let versions = install::list_installed_sop();
            if !versions.contains(&resolved) {
                if prompt_install(tool, &resolved)? {
                    install::install_sop(&resolved, false)?;
                } else {
                    return Ok(());
                }
            }

            config.default_sop = Some(resolved.clone());

            // Create/update symlink
            let target = paths::sop_binary(&resolved);
            let link = paths::sop_symlink();
            create_symlink(&target, &link)?;

            println!(
                "{} Default sop version set to {}",
                style("✓").green(),
                style(&resolved).bold()
            );

            // Auto-set compatible Go version
            if let Some(go_version) = find_or_install_compatible_go(&resolved)? {
                config.default_go = Some(go_version.clone());
                println!(
                    "{} Default go version set to {} (compatible with sop {})",
                    style("✓").green(),
                    style(&go_version).bold(),
                    &resolved
                );
            }

            config.save()?;
            print_path_hint();
        }
        _ => {
            return Err(format!("Unknown tool '{}'. Use 'go' or 'sop'.", tool).into());
        }
    }
    Ok(())
}

fn prompt_install(tool: &str, version: &str) -> Result<bool, Box<dyn std::error::Error>> {
    use std::io::{self, Write};

    print!(
        "{} {} is not installed. Install it? [Y/n] ",
        style(tool).bold(),
        style(version).bold()
    );
    io::stdout().flush()?;

    let mut input = String::new();
    io::stdin().read_line(&mut input)?;

    let input = input.trim().to_lowercase();
    Ok(input.is_empty() || input == "y" || input == "yes")
}

fn create_symlink(target: &std::path::Path, link: &std::path::Path) -> std::io::Result<()> {
    // Remove existing symlink/file if present
    if link.exists() || link.symlink_metadata().is_ok() {
        std::fs::remove_file(link)?;
    }

    #[cfg(unix)]
    {
        std::os::unix::fs::symlink(target, link)?;
    }

    #[cfg(windows)]
    {
        // On Windows, try symlink first, fall back to hard link or copy
        if std::os::windows::fs::symlink_file(target, link).is_err() {
            // Symlinks may require admin privileges on Windows
            // Fall back to copying the file
            std::fs::copy(target, link)?;
        }
    }

    Ok(())
}

fn print_path_hint() {
    let bin_dir = paths::bin_dir();

    // Check if bin_dir is likely in PATH already
    if let Ok(path_var) = std::env::var("PATH") {
        let bin_str = bin_dir.to_string_lossy();
        if path_var.contains(bin_str.as_ref()) {
            return; // Already in PATH, no hint needed
        }
    }

    println!();
    println!(
        "{} Add {} to your PATH:",
        style("hint:").cyan(),
        style("~/.sopmod/bin").bold()
    );
    println!();

    if cfg!(windows) {
        println!(
            "  {}",
            style("$env:PATH = \"$HOME\\.sopmod\\bin;$env:PATH\"").dim()
        );
    } else {
        println!(
            "  {}",
            style("export PATH=\"$HOME/.sopmod/bin:$PATH\"").dim()
        );
    }
}

fn cmd_which(tool: &str) -> Result<(), Box<dyn std::error::Error>> {
    let config = Config::load();

    match tool {
        "go" => {
            // Go versions are managed by sop, not exposed directly
            let versions = install::list_installed_go();
            if versions.is_empty() {
                return Err("No go versions installed. Run `sopmod install go latest`".into());
            } else {
                println!("{}", style("Installed go versions:").bold());
                for v in &versions {
                    let path = paths::go_binary(v);
                    println!("  {} -> {}", style(v).green(), path.display());
                }
                return Ok(());
            }
        }
        "sop" => {
            if let Some(version) = config.default_sop {
                let path = paths::sop_binary(&version);
                if path.exists() {
                    println!("{}", path.display());
                } else {
                    return Err(format!(
                        "sop {} is set as default but not found at {:?}",
                        version, path
                    )
                    .into());
                }
            } else {
                // No default set, check if any versions are installed
                let versions = install::list_installed_sop();
                if versions.is_empty() {
                    return Err("No sop versions installed. Run `sopmod install sop latest`".into());
                } else {
                    return Err(format!(
                        "No default sop version set. Run `sopmod default sop {}`",
                        versions[0]
                    )
                    .into());
                }
            }
        }
        _ => {
            return Err(format!("Unknown tool '{}'. Use 'go' or 'sop'.", tool).into());
        }
    }
    Ok(())
}

fn cmd_remove(tool: &str, version: &str) -> Result<(), Box<dyn std::error::Error>> {
    let mut config = Config::load();

    match tool {
        "go" => {
            let resolved = install::resolve_go_version(version)?;
            install::remove_go(&resolved)?;
        }
        "sop" => {
            let resolved = install::resolve_sop_version(version)?;
            install::remove_sop(&resolved)?;
            // Clear default if it was this version
            if config.default_sop.as_deref() == Some(&resolved) {
                config.default_sop = None;
                config.save()?;
            }
        }
        _ => {
            return Err(format!("Unknown tool '{}'. Use 'go' or 'sop'.", tool).into());
        }
    }
    Ok(())
}

fn cmd_update(tool: Option<&str>) -> Result<(), Box<dyn std::error::Error>> {
    match tool {
        Some("go") => update_go()?,
        Some("sop") => update_sop()?,
        None => {
            update_go()?;
            update_sop()?;
        }
        Some(other) => {
            return Err(format!("Unknown tool '{}'. Use 'go' or 'sop'.", other).into());
        }
    }
    Ok(())
}

fn update_go() -> Result<(), Box<dyn std::error::Error>> {
    // Resolve and install latest
    let latest = install::resolve_go_version("latest")?;
    let installed = install::list_installed_go();

    if installed.contains(&latest) {
        println!(
            "{} go {} is already the latest version",
            style("✓").green(),
            style(&latest).bold()
        );
    } else {
        install::install_go(&latest, false)?;
    }

    Ok(())
}

fn update_sop() -> Result<(), Box<dyn std::error::Error>> {
    let mut config = Config::load();
    let old_default = config.default_sop.clone();

    // Resolve and install latest
    let latest = install::resolve_sop_version("latest")?;
    let installed = install::list_installed_sop();

    if installed.contains(&latest) {
        println!(
            "{} sop {} is already the latest version",
            style("✓").green(),
            style(&latest).bold()
        );
    } else {
        install::install_sop(&latest, false)?;
    }

    // Update default only if current default matches an installed version
    // (meaning user was tracking latest, not pinned to a specific version)
    let should_update_default = match &old_default {
        None => true,
        Some(d) => installed.contains(d),
    };

    if should_update_default && config.default_sop.as_deref() != Some(&latest) {
        config.default_sop = Some(latest.clone());

        let target = paths::sop_binary(&latest);
        let link = paths::sop_symlink();
        create_symlink(&target, &link)?;

        println!(
            "{} Default sop version updated to {}",
            style("✓").green(),
            style(&latest).bold()
        );

        // Auto-update compatible Go version
        if let Some(go_version) = find_or_install_compatible_go(&latest)? {
            config.default_go = Some(go_version.clone());
            println!(
                "{} Default go version set to {} (compatible with sop {})",
                style("✓").green(),
                style(&go_version).bold(),
                &latest
            );
        }

        config.save()?;
    }

    Ok(())
}

/// Find the best compatible Go version for a sop version, installing if needed
fn find_or_install_compatible_go(
    sop_version: &str,
) -> Result<Option<String>, Box<dyn std::error::Error>> {
    let Some(compat_info) = compat::go_compat(sop_version) else {
        // Unknown sop version, no Go requirement known
        return Ok(None);
    };

    let installed_go = install::list_installed_go();

    // Find the best (latest) compatible installed Go version
    let best_installed = installed_go
        .iter()
        .filter(|v| compat::is_go_compatible(v, sop_version).unwrap_or(false))
        .max_by(|a, b| compare_versions(a, b));

    if let Some(version) = best_installed {
        return Ok(Some(version.clone()));
    }

    // No compatible Go installed, install the minimum required version
    println!(
        "{} Installing go {} (required by sop {})...",
        style("→").cyan(),
        style(compat_info.min).bold(),
        sop_version
    );
    let installed = install::install_go(compat_info.min, false)?;
    Ok(Some(installed))
}

/// Compare two version strings (semver-style)
fn compare_versions(a: &str, b: &str) -> std::cmp::Ordering {
    let parse = |v: &str| -> Vec<u32> { v.split('.').filter_map(|p| p.parse().ok()).collect() };
    parse(a).cmp(&parse(b))
}
