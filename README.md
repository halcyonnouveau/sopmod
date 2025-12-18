# SOPMOD

Version manager for [Soppo](https://github.com/halcyonnouveau/soppo).

SOPMOD handles installing, updating, and switching between multiple versions of Sop. It also manages Go installations that Sop uses internally for compilation.

## Install

```bash
curl -fsSL https://soppolang.dev/install.sh | sh
```

Or with cargo:

```bash
cargo install sopmod
```

After installation, add `~/.sopmod/bin` to your PATH:

```bash
export PATH="$HOME/.sopmod/bin:$PATH"
```

## Usage

```bash
# Install latest sop (automatically set as default)
sopmod install sop latest

# Install a specific version
sopmod install sop 0.4.1

# Update to latest
sopmod update sop

# Set default version
sopmod default sop 0.4.1

# List installed versions
sopmod list

# Show active binary path
sopmod which sop

# Remove a version
sopmod remove sop 0.4.0
```

## Per-project versions

Pin a specific Sop version for your project by adding to `sop.mod`:

```toml
sop = "0.4"
```

When you run `sop` in a directory with a `sop.mod` file, SOPMOD automatically uses the pinned version. Version matching is flexible - `sop = "0.4"` will match any 0.4.x version installed.

If the required version isn't installed, you'll see:

```
error: sop.mod requires sop 0.5, but 0.4.1 is installed

  hint: sopmod can manage multiple versions
```

## Go versions

Sop compiles to Go, so it needs a Go installation. SOPMOD manages this automatically:

```bash
# Install Go (usually not needed - done automatically)
sopmod install go 1.22

# List installed Go versions
sopmod list go
```

When you set a default Sop version, SOPMOD automatically installs and configures a compatible Go version. You can also pin Go in `sop.mod`:

```toml
go = "1.22"
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
