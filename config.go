package radvd_manager

import (
	"log"
	"os"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

const (
	templateFile         = "radvd.template.conf"
	RadvdConfPath        = "/etc/radvd.d/"
	debugConfPath        = "./output/"
	parameterFile        = "parameter.default.yaml"
	defaultRadvdCondFile = "/etc/radvd.conf"
)

type Policy struct {
	Rules  []Rule  `yaml:"rules" validate:"unique=ID,required,dive" default:"[]"`
	Groups []Group `yaml:"groups" validate:"unique=ID,required,dive" default:"[]"`
}

type Rule struct {
	ID          int    `yaml:"id" validate:"required"`
	Description string `yaml:"description"`
	// Type        string   `yaml:"type" validate:"oneof=FQDNs Prefixes,required"`
	// FQDNs       []string `yaml:"fqdn,omitempty" validate:"dive,domain"`
	Prefixes []string `yaml:"prefixes,omitempty" validate:"dive,cidrv6"`
	Nexthop  string   `yaml:"nexthop" validate:"ipv6,required"`
}

type Group struct {
	// ID          int      `yaml:"id" validate:"required"`
	Description string   `yaml:"description"`
	Rules       []int    `yaml:"rules" validate:"dive,chechk_rule_exist,required"`
	Members     []string `yaml:"members" validate:"dive,ipv6,required"`
}

func ParseConfig(policy *Policy) ([]*Instance, error) {
	instances := []*Instance{}
	parameters, err := LoadParameterFile()
	if err != nil {
		log.Fatalf("Failed to marshal radvd to JSON: %v", err)
	}
	for _, i := range policy.Rules {
		var new Instance
		for _, j := range parameters {
			if j.RouterID == i.Nexthop {
				new = *j
				break
			}
		}
		new.ID = uint32(i.ID)
		if !is_contain(i.Prefixes, "::/0") {
			for _, j := range i.Prefixes {
				route := Route{
					Route:              j,
					AdvRouteLifetime:   1800,
					AdvRoutePreference: "medium",
				}
				new.Routes = append(new.Routes, route)
			}
		} else if len(i.Prefixes) == 1 && i.Prefixes[0] == "::/0" {
			// "::/0" must be the only element in the list
			new.AdvDefaultPreference = "high"
		}
		instances = append(instances, &new)
	}

	for _, i := range policy.Groups {
		for _, j := range i.Rules {
			for _, k := range instances {
				if k.ID == uint32(j) {
					k.Clients = append(k.Clients, i.Members...)
				}
			}
		}
	}

	for _, i := range instances {
		err := GenerateRadvdConfigFile(i, debugConfPath)
		if err != nil {
			return nil, err
		}
	}

	return instances, nil
}

func LoadPolicyFile(filePath string) (*Policy, error) {
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var policy Policy
	if err = yaml.Unmarshal(fileData, &policy); err != nil {
		return nil, err
	}
	// validate the rules
	for _, i := range policy.Rules {
		if err := policy.validateRule(i); err != nil {
			return nil, err
		}
	}
	// validate the group
	for _, i := range policy.Groups {
		if err := policy.validateGroup(i); err != nil {
			return nil, err
		}
	}

	return &policy, nil
}

func LoadParameterFile() ([]*Instance, error) {
	fileData, err := os.ReadFile(parameterFile)
	if err != nil {
		return nil, err
	}
	instances := []*Instance{}
	if err = yaml.Unmarshal(fileData, instances); err != nil {
		return nil, err
	}

	return instances, nil
}

type ValidationErrors = validator.ValidationErrors

//var domainRegexp = regexp.MustCompile(`^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]$`)

func (c *Policy) validateRule(rule Rule) error {
	// validate := validator.New(validator.WithRequiredStructEnabled())

	// TODO:

	return nil
}

func (c *Policy) validateGroup(group Group) error {
	validate := validator.New(validator.WithRequiredStructEnabled())

	validate.RegisterValidation("chechk_rule_exist", func(fl validator.FieldLevel) bool {

		for _, i := range c.Rules {
			for _, j := range group.Rules {
				if i.ID == j {
					return true
				}
			}
		}
		return false
	})

	// TODO:

	return nil
}

func is_contain(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
