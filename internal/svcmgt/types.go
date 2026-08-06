// Package svcmgt provides systemd service management commands for the server CLI.
// It supports listing, inspecting, starting, stopping, restarting, enabling,
// disabling, and viewing logs for systemd services using systemctl and journalctl.
package svcmgt

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// Scope determines whether to manage system or user systemd services.
type Scope string

const (
	ScopeSystem Scope = "system"
	ScopeUser   Scope = "user"
)

// scopeFlag returns the systemctl flag for the given scope.
func scopeFlag(scope Scope) string {
	if scope == ScopeUser {
		return "--user"
	}
	return "--system"
}

// scopeFlagArgs returns the scope flag as a string slice for exec.Command.
func scopeFlagArgs(scope Scope) []string {
	return []string{scopeFlag(scope)}
}

// ServiceState represents the active state of a systemd unit.
type ServiceState string

const (
	StateActive     ServiceState = "active"
	StateInactive   ServiceState = "inactive"
	StateFailed     ServiceState = "failed"
	StateActivating ServiceState = "activating"
	StateUnknown    ServiceState = "unknown"
)

// Service describes a systemd service unit.
type Service struct {
	Unit        string       `json:"unit"`
	LoadState   string       `json:"load_state"`
	ActiveState ServiceState `json:"active_state"`
	SubState    string       `json:"sub_state"`
	Description  string       `json:"description"`
	Enabled     bool         `json:"enabled"`
	Type        string       `json:"type"`
	PID         int          `json:"pid"`
}

// ServiceDetail holds extended information about a single service.
type ServiceDetail struct {
	Service
	FragmentPath string            `json:"fragment_path"`
	ExecStart     string            `json:"exec_start"`
	ExecMainPID   int               `json:"exec_main_pid"`
	MemoryCurrent uint64            `json:"memory_current"`
	MemoryPeak    uint64            `json:"memory_peak"`
	CPUUsageNS    uint64            `json:"cpu_usage_ns"`
	Timestamp     string            `json:"timestamp"`
	Environment   map[string]string `json:"environment,omitempty"`
}

// GatherServices runs systemctl list-units and returns all service units.
func GatherServices(scope Scope) ([]Service, error) {
	args := append(scopeFlagArgs(scope), "list-units", "--type=service",
		"--all", "--no-legend", "--no-pager",
		"--output=json")
	out, err := exec.Command("systemctl", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("running systemctl list-units: %w", err)
	}

	var units []struct {
		Unit        string `json:"unit"`
		LoadState   string `json:"load"`
		ActiveState string `json:"active"`
		SubState    string `json:"sub"`
		Description string `json:"description"`
	}

	if err := json.Unmarshal(out, &units); err != nil {
		return nil, fmt.Errorf("parsing systemctl JSON: %w", err)
	}

	// Gather enabled state for all services in one call
	enabledMap := gatherEnabledMap(scope)

	var result []Service
	for _, u := range units {
		result = append(result, Service{
			Unit:        u.Unit,
			LoadState:   u.LoadState,
			ActiveState: ServiceState(u.ActiveState),
			SubState:    u.SubState,
			Description: u.Description,
			Enabled:     enabledMap[u.Unit],
			Type:        "service",
		})
	}

	return result, nil
}

// gatherEnabledMap returns a set of enabled service unit names.
func gatherEnabledMap(scope Scope) map[string]bool {
	args := append(scopeFlagArgs(scope), "list-unit-files", "--type=service",
		"--no-legend", "--no-pager")
	out, err := exec.Command("systemctl", args...).Output()
	if err != nil {
		return nil
	}

	result := make(map[string]bool)
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			result[fields[0]] = fields[1] == "enabled"
		}
	}
	return result
}

// FindService returns the service with the given unit name, or nil.
func FindService(services []Service, unit string) *Service {
	for i := range services {
		if services[i].Unit == unit {
			return &services[i]
		}
	}
	return nil
}

// GatherServiceDetail runs systemctl show on a single service.
func GatherServiceDetail(unit string, scope Scope) (*ServiceDetail, error) {
	args := append(scopeFlagArgs(scope), "show", unit, "--no-pager")
	out, err := exec.Command("systemctl", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("running systemctl show %s: %w", unit, err)
	}

	props := parseSystemctlShow(string(out))
	if len(props) == 0 {
		return nil, fmt.Errorf("service %q not found", unit)
	}

	detail := &ServiceDetail{
		Service: Service{
			Unit:        unit,
			LoadState:   props["LoadState"],
			ActiveState: ServiceState(props["ActiveState"]),
			SubState:    props["SubState"],
			Description: props["Description"],
		},
		FragmentPath: props["FragmentPath"],
		ExecStart:    props["ExecStart"],
	}

	if detail.LoadState == "loaded" || detail.LoadState == "masked" {
		// Determine enabled state
		if state := props["UnitFileState"]; state == "enabled" || state == "static" {
			detail.Enabled = state == "enabled"
		}
	}

	if pid := props["MainPID"]; pid != "" {
		fmt.Sscanf(pid, "%d", &detail.PID)
		fmt.Sscanf(pid, "%d", &detail.ExecMainPID)
	}

	if v := props["ExecMainPID"]; v != "" {
		fmt.Sscanf(v, "%d", &detail.ExecMainPID)
	}

	if v := props["MemoryCurrent"]; v != "" && v != "[not set]" {
		fmt.Sscanf(v, "%d", &detail.MemoryCurrent)
	}
	if v := props["MemoryPeak"]; v != "" && v != "[not set]" {
		fmt.Sscanf(v, "%d", &detail.MemoryPeak)
	}
	if v := props["CPUUsageNSec"]; v != "" && v != "[not set]" {
		fmt.Sscanf(v, "%d", &detail.CPUUsageNS)
	}

	detail.Timestamp = props["ActiveEnterTimestamp"]
	if detail.Timestamp == "" {
		detail.Timestamp = props["StateChangeTimestamp"]
	}

	if env := props["Environment"]; env != "" {
		detail.Environment = make(map[string]string)
		for _, kv := range strings.Fields(env) {
			if idx := strings.Index(kv, "="); idx > 0 {
				detail.Environment[kv[:idx]] = kv[idx+1:]
			}
		}
	}

	return detail, nil
}

// parseSystemctlShow parses the key=value output of systemctl show.
func parseSystemctlShow(output string) map[string]string {
	result := make(map[string]string)
	for _, line := range strings.Split(output, "\n") {
		idx := strings.Index(line, "=")
		if idx <= 0 {
			continue
		}
		result[line[:idx]] = line[idx+1:]
	}
	return result
}

// ServiceAction performs a systemctl action on a service.
func ServiceAction(unit, action string, scope Scope) error {
	args := append(scopeFlagArgs(scope), action, unit)
	cmd := exec.Command("systemctl", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("systemctl %s %s: %w\n%s", action, unit, err, string(out))
	}
	return nil
}

// StartService starts a systemd service.
func StartService(unit string, scope Scope) error {
	return ServiceAction(unit, "start", scope)
}

// StopService stops a systemd service.
func StopService(unit string, scope Scope) error {
	return ServiceAction(unit, "stop", scope)
}

// RestartService restarts a systemd service.
func RestartService(unit string, scope Scope) error {
	return ServiceAction(unit, "restart", scope)
}

// EnableService enables a systemd service to start at boot.
func EnableService(unit string, scope Scope) error {
	return ServiceAction(unit, "enable", scope)
}

// DisableService disables a systemd service from starting at boot.
func DisableService(unit string, scope Scope) error {
	return ServiceAction(unit, "disable", scope)
}

// ReloadService reloads a systemd service's configuration.
func ReloadService(unit string, scope Scope) error {
	return ServiceAction(unit, "reload", scope)
}

// MaskService masks a systemd service so it cannot be started.
func MaskService(unit string, scope Scope) error {
	return ServiceAction(unit, "mask", scope)
}

// UnmaskService unmasks a previously masked service.
func UnmaskService(unit string, scope Scope) error {
	return ServiceAction(unit, "unmask", scope)
}

// GatherLogs runs journalctl for the given service unit.
func GatherLogs(unit string, lines int, follow bool, scope Scope) (string, error) {
	args := []string{"-u", unit, "--no-pager"}
	if scope == ScopeUser {
		args = append([]string{"--user"}, args...)
	}
	if lines > 0 {
		args = append(args, "-n", fmt.Sprintf("%d", lines))
	} else {
		args = append(args, "-n", "50")
	}
	if follow {
		args = append(args, "-f")
	}

	out, err := exec.Command("journalctl", args...).Output()
	if err != nil {
		return "", fmt.Errorf("running journalctl for %s: %w", unit, err)
	}
	return string(out), nil
}