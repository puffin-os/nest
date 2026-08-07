# Changelog

## [1.2.0](https://github.com/puffin-os/nest/compare/v1.1.1...v1.2.0) (2026-08-07)


### Features

* add update subcommand to all CLIs ([f21c2cf](https://github.com/puffin-os/nest/commit/f21c2cfebc57687a594f97a28e3e3cc2720b97ca))

## [1.1.1](https://github.com/puffin-os/nest/compare/v1.1.0...v1.1.1) (2026-08-07)


### Bug Fixes

* use nest as command name in help output ([8c24354](https://github.com/puffin-os/nest/commit/8c24354c8ef2c34cae2e7989d271c37b1cc479bb))

## [1.1.0](https://github.com/puffin-os/nest/compare/v1.0.0...v1.1.0) (2026-08-07)


### Features

* add --user flag for user-level systemd services ([56dbe62](https://github.com/puffin-os/nest/commit/56dbe620d2a766c35ecde337803fb96da3d16ace))
* add apps (quadlet) subcommand to nest-server ([911fe19](https://github.com/puffin-os/nest/commit/911fe19ecb39a88614e8fd9f5603a5ed9aa6b65b))
* add apps inspect command for quadlet runtime stats ([01132d3](https://github.com/puffin-os/nest/commit/01132d387fc4c35e433d52d8782853adca5e4074))
* add disk management subcommands to nest-server ([41b49ab](https://github.com/puffin-os/nest/commit/41b49ab46d111549de181eb2b7ad42d7af9c38cb))
* add logs subcommand to apps (quadlet) ([84c8716](https://github.com/puffin-os/nest/commit/84c87160838cf44dc7fafecb4fd1e9c14961e30e))
* add network and disk subcommands to nest-workstation ([e681ed2](https://github.com/puffin-os/nest/commit/e681ed2bd9878bc47a3bc9d7073c4e4865deea66))
* add network and restart selection to quadlet create wizard ([6ef2179](https://github.com/puffin-os/nest/commit/6ef2179e14b4f901dadb68bd696b0d73f36e3913))
* add service subcommand to nest-workstation ([fdd7417](https://github.com/puffin-os/nest/commit/fdd741781c53d4f250b6df94ffcef8a99be31a26))
* add systemd service management subcommands to nest-server ([02a78c5](https://github.com/puffin-os/nest/commit/02a78c53e81d4a0be11f3e707568b96bf647270e))


### Bug Fixes

* don't pass --user flag to podman in inspect ([b4682cf](https://github.com/puffin-os/nest/commit/b4682cf5624425d54bb340cb9595e89688076b9c))
* journalctl -f follow mode now works for apps and service logs ([363f41f](https://github.com/puffin-os/nest/commit/363f41ff4c26d793bbc2323e68d12b54c23d4cb6))
* use correct .service suffix for quadlet systemd units ([feef85c](https://github.com/puffin-os/nest/commit/feef85c7394dec51524b15be4202bfedfa925b56))

## 1.0.0 (2026-08-06)


### Features

* 2-column layout with single frame for system-info ([513c4a8](https://github.com/puffin-os/nest/commit/513c4a83561f0a8f2cb5966f3e5f7d6f270e1b89))
* add flavor-specific command registration mechanism ([66ef486](https://github.com/puffin-os/nest/commit/66ef48609dabc42669460e5e6578e39e3ea1163f))
* add lipgloss styled output for system-info ([8a15183](https://github.com/puffin-os/nest/commit/8a151831da60804ee12ad99693a991523056a4cc))
* add server-only network management subcommand ([26d2ed3](https://github.com/puffin-os/nest/commit/26d2ed33b9c54b0d33fb3d6afd00b185ffcafcae))
* initial nest CLI skeleton with system-info subcommand ([5f07aef](https://github.com/puffin-os/nest/commit/5f07aefc515716765dd4e4ba2f9f8bce2de9ef42))
* split system-info into 3 grouped sections ([2592560](https://github.com/puffin-os/nest/commit/2592560e13c858f2167794014e845132899396ab))
