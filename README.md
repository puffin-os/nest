# nest

Shared Go CLI for Puffin OS derivatives. One codebase, three binaries:

- **nest-server** — server derivative
- **nest-desktop** — desktop derivative
- **nest-workstation** — workstation derivative

## Build

```sh
make all        # builds all three binaries into bin/
make nest-server # build a single binary
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
  sysinfo/           # system info gathering (gopsutil)
```

## Libraries

- [cobra](https://github.com/spf13/cobra) — CLI framework
- [gopsutil](https://github.com/shirou/gopsutil) — cross-platform system metrics