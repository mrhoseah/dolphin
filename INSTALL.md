# Installing Dolphin CLI

## Local Installation

For local development, install from the dolphin directory:

```bash
cd /path/to/dolphin
go install .
```

This installs the CLI to `$GOPATH/bin/dolphin` (usually `~/go/bin/dolphin`).

## Using Local Binary

Alternatively, you can build and use the binary directly:

```bash
cd /path/to/dolphin
go build -o dolphin ./main.go

# Then use it directly
./dolphin make:auth

# Or from another project
../dolphin/dolphin make:auth
```

## Troubleshooting

### Error: "malformed module path"

If you see:
```
go: dolphin/cmd/dolphin@latest: malformed module path "dolphin/cmd/dolphin": missing dot in first path element
```

This means you tried to install using `go install dolphin/cmd/dolphin@latest`, which won't work because "dolphin" is not a valid module path (it needs a domain like `github.com/user/dolphin`).

**Solution**: Install locally using `go install .` from the dolphin directory instead.

### Module Path Requirements

Go module paths must have a dot in the first path element (like `github.com/user/project` or `example.com/project`). The current module is just `dolphin` for local development, so you can't install it via `go install` with a module path.

## Verifying Installation

Check if dolphin is installed:

```bash
which dolphin
dolphin --help
```

## Updating Installation

To update the global installation:

```bash
cd /path/to/dolphin
go install .
```

