# nest

Shared Go CLI for Puffin OS derivatives. One codebase, three binaries:

- **nest-server** — server derivative (includes network management)
- **nest-desktop** — desktop derivative
- **nest-workstation** — workstation derivative

## Architecture

The CLI uses [cobra](https://github.com/spf13/cobra) for command structure and
[lipgloss](https://github.com/charmbracelet/lipgloss) for styled terminal output.
Interactive forms use [bubbletea](https://github.com/charmbracelet/bubbletea) and
[bubbles](https://github.com/charmbracelet/bubbles). System metrics are gathered
with [gopsutil](https://github.com/shirou/gopsutil).

### Shared codebase, per-flavor capabilities

All three binaries share the same Go module. Each binary entrypoint lives in
`cmd/nest-<flavor>/main.go` and blank-imports its flavor package from
`internal/cli/<flavor>/`. The flavor package registers subcommands via
`cli.RegisterFlavorCmds` in an `init()`.

To add a shared subcommand: add it in `internal/cli/cli.go`.
To add a server-only subcommand: add it in `internal/cli/server/server.go`.

### Directory layout

```
cmd/
  nest-server/        # entrypoint, blank-imports internal/cli/server
  nest-desktop/       # entrypoint, blank-imports internal/cli/desktop
  nest-workstation/   # entrypoint, blank-imports internal/cli/workstation
internal/
  cli/                # shared cobra command tree, RegisterFlavorCmds hook
    server/           # server-only subcommands (network)
    desktop/          # desktop-only subcommands (placeholder)
    workstation/      # workstation-only subcommands (placeholder)
  netmgmt/            # network management: list, status, add, remove
  diskmgmt/           # disk management: list, status, mounts, mount, unmount, format, expand
  svcmgt/             # service management: list, status, start/stop/restart, enable/disable, logs
  sysinfo/            # system info gathering and formatted output
.github/workflows/    # CI (build/vet/test) and release-please
```

## Commands

### Shared

- `system-info` — prints CPU, memory, disk, network, OS, and Go runtime info
  - `--json` for machine-readable output
  - `--plain` for unstyled text (default is lipgloss styled 2-column layout)

### Server-only

- `network` — manage network interfaces, VLANs, and bonds
  - `network list` — table of all interfaces (`--json` supported)
  - `network status <iface>` — detailed view with addresses, stats, bond/VLAN info
  - `network add` — interactive bubbletea form for VLAN/bond/dummy (`--dry-run`)
  - `network remove` — interactive selection list (`--dry-run`)

- `disk` — manage block devices, partitions, and filesystems
  - `disk list` — table of all block devices (`--json` supported)
  - `disk status <device>` — detailed view with partitions, filesystem, mount info
  - `disk mounts` — table of mounted filesystems with size/used/avail/use% (`--json`)
  - `disk mount <device> <mountpoint>` — mount a device (`-o` for mount options)
  - `disk unmount <device|mountpoint>` — unmount (`--lazy` for lazy unmount)
  - `disk format <device>` — format with filesystem (`-t ext4|xfs|btrfs|fat32|swap`, `-l` for label)
  - `disk expand <device>` — grow partition and resize filesystem to fill available space

- `service` (alias `svc`) — manage systemd services
  - `service list` — table of all services (`--json`, `--state` to filter)
  - `service status <service>` — detailed view with PID, memory, CPU, environment (`--json`)
  - `service start <service>` — start a service
  - `service stop <service>` — stop a service
  - `service restart <service>` — restart a service
  - `service reload <service>` — reload service configuration
  - `service enable <service>` — enable at boot
  - `service disable <service>` — disable at boot
  - `service mask <service>` — mask so it cannot be started
  - `service unmask <service>` — unmask a previously masked service
  - `service logs <service>` — view journal logs (`-n` for lines, `-f` to follow)

## Build and development

Uses [Taskfile](https://taskfile.dev) (not Makefile):

```sh
task build              # build all three binaries into bin/
task build:server       # build a single binary
task vet                # run go vet
task test               # run go test
task tidy               # go mod tidy
task clean              # remove bin/
```

## Release

Versioning is automated via [release-please](https://github.com/googleapis/release-please).
Use [conventional commits](https://www.conventionalcommits.org/) (`feat:`, `fix:`, `deps:`, etc.)
for all commits. release-please creates a release PR on pushes to `main` and tags releases.

Config: `release-please-config.json`, manifest: `.release-please-manifest.json`.

## CI

GitHub Actions workflows in `.github/workflows/`:

- `ci.yml` — runs `go vet`, builds all three binaries, `go test` on push/PR
- `release.yml` — runs release-please on push to main

## Invariants

- Go module path: `github.com/puffin/nest`
- Keep the base small. A dependency belongs in the first binary that needs it.
- Flavor-specific subcommands go in `internal/cli/<flavor>/`, not in `internal/cli/`.
- Interactive commands (bubbletea) must degrade gracefully without a TTY.
- All output modes (styled, plain, JSON) must be tested when adding or modifying a command.