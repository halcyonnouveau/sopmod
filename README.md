# SOPMOD

Version manager for [Soppo](https://github.com/halcyonnouveau/soppo).

## Install

```bash
curl -fsSL https://soppolang.dev/install.sh | sh
```

Or with cargo:

```bash
cargo install sopmod
```

## Usage

```bash
# Install sop and go
sopmod install sop latest
sopmod install go latest

# Set default (creates symlink in ~/.sopmod/bin/)
sopmod default sop latest

# Update to latest
sopmod update sop

# List installed versions
sopmod list

# Show active binary path
sopmod which sop

# Remove a version
sopmod remove sop 0.4.0
```

Add `~/.sopmod/bin` to your PATH:

```bash
export PATH="$HOME/.sopmod/bin:$PATH"
```

## Per-project versions

Add to your `sop.mod`:

```toml
sop = "0.4.0"
```

The `sop` binary will use the specified version from `~/.sopmod/` when available.

## Go versions

SOPMOD also manages Go versions that `sop` uses internally:

```bash
sopmod install go 1.22
```

A compatible Go version is automatically set when you run `sopmod default sop <version>`.
