/**
 * Manage radvd service.
 * Start, stop, reload radvd service.
 *
 */
package radvd_manager

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// = Interface
type Instance struct {
	// Metadata
	ID       uint32 `json:"id" yaml:"id"`
	PID      uint32 `json:"pid" yaml:"pid"`
	RouterID string `json:"router_id" yaml:"router_id"`
	Name     string `json:"name" yaml:"name"`
	// Configuration parameters for radvd
	AdvSendAdvert        bool     `json:"adv_send_advert" yaml:"adv_send_advert"`
	MinRtrAdvInterval    uint32   `json:"min_rtr_adv_interval" yaml:"min_rtr_adv_interval"`
	MaxRtrAdvInterval    uint32   `json:"max_rtr_adv_interval" yaml:"max_rtr_adv_interval"`
	AdvManagedFlag       bool     `json:"adv_managed_flag" yaml:"adv_managed_flag"`
	AdvOtherConfigFlag   bool     `json:"adv_other_config_flag" yaml:"adv_other_config_flag"`
	AdvDefaultLifetime   uint32   `json:"adv_default_lifetime" yaml:"adv_default_lifetime"`
	AdvDefaultPreference string   `json:"adv_default_preference" yaml:"adv_default_preference"`
	Prefixes             []Prefix `json:"prefixes" yaml:"prefixes"`
	Rdnss                []RDNSS  `json:"rdnss" yaml:"rdnss"`
	Routes               []Route  `json:"routes" yaml:"routes"`
	Clients              []string `json:"clients" yaml:"clients"`
}

type Prefix struct {
	Prefix           string `json:"prefix" yaml:"prefix"`
	AdvOnLink        bool   `json:"adv_on_link" yaml:"adv_on_link"`
	AdvAutonomous    bool   `json:"adv_autonomous" yaml:"adv_autonomous"`
	AdvRouterAddr    bool   `json:"adv_router_addr" yaml:"adv_router_addr"`
	AdvValidLifetime uint32 `json:"adv_valid_lifetime" yaml:"adv_valid_lifetime"`
}

type RDNSS struct {
	Address          string `json:"address" yaml:"address"`
	AdvRdnssLifetime uint32 `json:"adv_rdnss_lifetime" yaml:"adv_rdnss_lifetime"`
}

type Route struct {
	Route              string `json:"route" yaml:"route"`
	AdvRouteLifetime   uint32 `json:"adv_route_lifetime" yaml:"adv_route_lifetime"`
	AdvRoutePreference string `json:"adv_route_preference" yaml:"adv_route_preference"`
}

func CheckRadvdConfig(id int) error {
	cmd := exec.Command(
		"/usr/sbin/radvd",
		"-C", fmt.Sprintf("/etc/radvd.d/%d.conf", id),
		"--configtest",
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failt to check configure: %w", err)
	}
	return nil
}

func StartRadvd(id int) error {
	pidFile := "/var/run/radvd/radvd." + strconv.Itoa(id) + ".pid"
	cfgFile := "/etc/radvd.d/" + strconv.Itoa(id) + ".conf"

	cmd := exec.Command(
		"/usr/sbin/radvd",
		"-C", cfgFile,
		"-p", pidFile,
	)

	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		os.Remove(cfgFile)
		return fmt.Errorf("failed to start radvd: %w", err)
	}

	return nil
}

func ReloadRadvd(id int) error {
	pidFile := "/var/run/radvd/radvd." + strconv.Itoa(id) + ".pid"
	pidStr, err := os.ReadFile(pidFile)
	if err != nil {
		return fmt.Errorf("failed to read PID file: %w", err)
	}
	pid, err := strconv.Atoi(string(pidStr))
	if err != nil {
		fmt.Println("Error:", err)
		return fmt.Errorf("failed to convert PID to int: %w", err)
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find radvd process: %w", err)
	}
	if err := process.Signal(syscall.SIGHUP); err != nil {
		return fmt.Errorf("failed to reload radvd: %w", err)
	}

	return nil
}

func StopRadvd(id int) error {
	if id == 0 {
		return nil
	}
	pidFile := "/var/run/radvd/radvd." + strconv.Itoa(id) + ".pid"
	pidStr, err := os.ReadFile(pidFile)
	if err != nil {
		return fmt.Errorf("error opening PID file: %w", err)
	}

	pid, err := strconv.Atoi(string(pidStr))
	if err != nil {
		fmt.Println("Error:", err)
		return fmt.Errorf("failed to convert PID to int: %w", err)
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find radvd process: %w", err)
	}
	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to reload radvd: %w", err)
	}
	if err := os.Remove(pidFile); err != nil {
		return fmt.Errorf("faild to remove PID file: %w", err)
	}
	if err := os.Remove("/etc/radvd.d/" + strconv.Itoa(id) + ".conf"); err != nil {
		return fmt.Errorf("faild to remove config file: %w", err)
	}

	return nil
}

func GetRadvdPID(id int) (int, error) {
	pidFile := "/var/run/radvd/radvd." + strconv.Itoa(id) + ".pid"
	pidStr, err := os.ReadFile(pidFile)
	if err != nil {
		return 0, fmt.Errorf("failed to read PID file: %w", err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(pidStr)))
	if err != nil {
		fmt.Println("Error:", err)
		return 0, fmt.Errorf("failed to convert PID to int: %w", err)
	}

	return pid, nil
}
