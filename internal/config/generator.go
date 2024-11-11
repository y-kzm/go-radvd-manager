/**
 * Generate radvd configuration file from radvd.Interface struct using template.
 *
 */
package config

import (
	"fmt"
	"os"
	"strconv"
	"text/template"

	"github.com/y-kzm/go-radvd-manager/internal/radvd"
)

const (
	templatePath = "../../configs/template/radvd_template.conf"
)

func GenerateRadvdConfigFile(iface *radvd.Interface) error {
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	instanceStr := strconv.Itoa(int(iface.Instance))
	file, err := os.Create("/etc/radvd.d/" + instanceStr + ".conf")
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	err = tmpl.Execute(file, iface)
	if err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	return nil
}
