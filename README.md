# SOPMOD II

Version manager for [Soppo](https://github.com/halcyonnouveau/soppo). SOPMOD is entirely dogfooded and has been rewritten from Rust to Soppo, hence the "II".

SOPMOD handles installing, updating, and switching between multiple versions of Soppo. It also manages Go installations that Soppo uses internally for compilation.

## Installation

```bash
curl -fsSL https://soppolang.dev/install.sh | sh
```

After installation, add `~/.sopmod/bin` to your PATH:

```bash
export PATH="$HOME/.sopmod/bin:$PATH"
```

## Usage

```bash
# Install latest sop 
sopmod install sop latest

# Install a specific version
sopmod install sop 0.4.1

# Update to latest
sopmod update sop

# Set default version
sopmod default 0.4.1

# List installed versions
sopmod list

# Remove a version
sopmod remove sop 0.4.0
```

### Per-project versions

Pin a specific Soppo version for your project by adding to `sop.mod`:

```toml
sop = "0.4"
```

When you run `sop` in a directory with a `sop.mod` file, SOPMOD automatically uses the pinned version. Version matching is flexible - `sop = "0.4"` will match any 0.4.x version installed.

### Go versions

Soppo compiles to Go, so it needs a Go installation. SOPMOD manages this automatically and when you set a default Soppo version, SOPMOD automatically installs and configures a compatible Go version. You can also pin a Go version in `sop.mod`:

```toml
go = "1.25"
```

## How it works

SOPMOD installs versions to `~/.sopmod/`:

```
~/.sopmod/
  config.toml        # Default versions
  bin/
    sop              # Shim that dispatches to correct version
  go/
    1.22.0/
    1.23.0/
  sop/
    0.4.0/
    0.4.1/
```

The `sop` binary in `~/.sopmod/bin/` is a shim that:
1. Checks for `sop.mod` in the current or parent directories
2. Uses the pinned version if specified
3. Falls back to the default version from `config.toml`
4. Executes the appropriate `sop` binary

## Licence

BSD 3-Clause. See [LICENCE](LICENCE).
