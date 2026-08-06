# nest

Shared Go CLI for Puffin OS derivatives. One codebase, three binaries:

- **nest-server** — server derivative
- **nest-desktop** — desktop derivative
- **nest-workstation** — workstation derivative

## Build

```sh
task build              # builds all three binaries into bin/
task build:server       # build a single binary
```

## Usage

```sh
nest-server system-info           # human-readable system info
nest-server system-info --json    # JSON output
```

## Structure

```
cmd/
  nest-server/       # entrypoint for server binary
  nest-desktop/      # entrypoint for desktop binary
  nest-workstation/  # entrypoint for workstation binary
internal/
  cli/               # shared cobra command tree
    server/          # server-specific subcommands (registered via init)
    desktop/         # desktop-specific subcommands (registered via init)
    workstation/     # workstation-specific subcommands (registered via init)
  sysinfo/           # system info gathering (gopsutil)
```

## Adding Flavor-Specific Subcommands

Each binary blank-imports its flavor package, which registers subcommands
via `cli.RegisterFlavorCmds` in an `init()`. To add a server-only command,
add it in `internal/cli/server/server.go`. Shared commands go in
`internal/cli/cli.go`.

## Libraries

- [cobra](https://github.com/spf13/cobra) — CLI framework
- [gopsutil](https://github.com/shirou/gopsutil) — cross-platform system metrics