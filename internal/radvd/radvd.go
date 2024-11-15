/**
 * Manage radvd service.
 * Start, stop, reload radvd service.
 *
 */
package radvd

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"
)

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

type Interface struct {
	Instance             uint32   `json:"instance" yaml:"instance"`
	Nexthop              string   `json:"nexthop" yaml:"nexthop"`
	Name                 string   `json:"name" yaml:"name"`
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

type Radvd struct {
	Interfaces []*Interface `json:"interfaces"`
}

func CheckRadvdConfig(instance int) error {
	cmd := exec.Command(
		"/usr/sbin/radvd",
		"-C", fmt.Sprintf("/etc/radvd.d/%d.conf", instance),
		"--configtest",
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("config test failed: %w", err)
	}
	return nil
}

func StartRadvd(instance int) error {
	pidFile := "/var/run/radvd/radvd." + strconv.Itoa(instance) + ".pid"
	cfgFile := "/etc/radvd.d/" + strconv.Itoa(instance) + ".conf"

	cmd := exec.Command(
		"/usr/sbin/radvd",
		"-C", cfgFile,
		"-p", pidFile,
		"-m", "syslog",
	)

	/* for debug */
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		os.Remove(cfgFile)
		return fmt.Errorf("failed to start radvd: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		os.Remove(cfgFile)
		return fmt.Errorf("radvd process failed: %w, stdout: %s, stderr: %s", err, out.String(), stderr.String())
	}

	return nil
}

func ReloadRadvd(instance int) error {
	pidFile := "/var/run/radvd/radvd." + strconv.Itoa(instance) + ".pid"
	pidData, err := os.ReadFile(pidFile)
	if err != nil {
		return fmt.Errorf("failed to read PID file: %w", err)
	}

	var pid int
	fmt.Sscanf(string(pidData), "%d", &pid)

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find radvd process: %w", err)
	}

	if err := process.Signal(syscall.SIGHUP); err != nil {
		return fmt.Errorf("failed to reload radvd: %w", err)
	}

	return nil
}

func StopRadvd(instance int) error {
	pidFile := "/var/run/radvd/radvd." + strconv.Itoa(instance) + ".pid"
	file, err := os.Open(pidFile)
	if err != nil {
		return fmt.Errorf("error opening PID file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("error reading PID file: %w", err)
		}
		return fmt.Errorf("PID file is empty")
	}

	pid, err := strconv.Atoi(scanner.Text())
	if err != nil {
		return fmt.Errorf("error converting PID: %w", err)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("error finding radvd process: %w", err)
	}

	// Signal radvd process to terminate
	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("error stopping radvd: %w", err)
	}

	// Remove PID file after successful termination
	if err := os.Remove(pidFile); err != nil {
		return fmt.Errorf("error removing PID file: %w", err)
	}

	if err := os.Remove("/etc/radvd.d/" + strconv.Itoa(instance) + ".conf"); err != nil {
		return fmt.Errorf("error removing config file: %w", err)
	}

	return nil
}
