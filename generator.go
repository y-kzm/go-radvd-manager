package radvd_manager

import (
	"fmt"
	"os"
	"strconv"
	"text/template"
)

func GenerateRadvdConfigFile(i *Instance, filePath string) error {
	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}
	conf, err := os.Create("/etc/radvd.d/" + strconv.Itoa(int(i.ID)) + ".conf")
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer conf.Close()
	if err = tmpl.Execute(conf, i); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	return nil
}
