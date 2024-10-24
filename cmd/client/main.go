package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/y-kzm/go-radvd-manager/internal/client"
	"github.com/y-kzm/go-radvd-manager/internal/config"
)

const (
	port = 8888
)

func main() {
	methodFlag := flag.String("X", "", "HTTP Method (POST, GET, DELETE)")
	fileFlag := flag.String("f", "", "Policy config file")
	flag.Parse()

	// read policy config file
	cfg, err := config.LoadPolicyConfig(*fileFlag)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// print config file content
	fmt.Println("--- Rules: ---")
	for _, rule := range cfg.Rules {
		fmt.Printf("ID: %d, Type: %s, FQDN: %s, Prefix: %s, Nexthop: %s\n",
			rule.ID, rule.Type, rule.FQDNs, rule.Prefixes, rule.Nexthop)
	}
	fmt.Println("--- Policies: ---")
	for _, policy := range cfg.Policies {
		fmt.Printf("ID: %d, Description: %s, Rules: %v, Clients: %v\n",
			policy.ID, policy.Description, policy.Rules, policy.Clients)
	}

	// generate radvd config files from policy config
	radvdConfigs, err := cfg.GenerateRadvdConfigFile()
	if err != nil {
		log.Fatalf("Failed to generate radvd config files: %v", err)
	}
	fmt.Println("\nRadvd config files generated successfully")

	// create client
	var clients []*client.RadvdManagerClient
	severs := client.GetSeverList(radvdConfigs)
	for _, server := range severs {
		client := client.NewClient(fmt.Sprintf("http://[%s]:%d", server, port), server, port)
		clients = append(clients, client)
	}

	if *fileFlag == "" || *methodFlag == "" {
		log.Fatal("Usage: -X POST -f <config_file>")
		return
	}

	var wg sync.WaitGroup

	// send request
	switch *methodFlag {
	case "POST":
		for _, radvdConfig := range radvdConfigs {
			file, err := os.ReadFile(radvdConfig.FilePath)
			if err != nil {
				log.Fatalf("Failed to read config file: %v", err)
			}
			for _, policy := range cfg.Policies {
				if config.ContainsInt(policy.Rules, radvdConfig.Rule.ID) {
					for _, c := range clients {
						if c.Server == radvdConfig.Rule.Nexthop {
							wg.Add(1)
							go func(client *client.RadvdManagerClient, id int, file string) {
								defer wg.Done()
								if err := client.Create(id, file); err != nil {
									log.Fatalf("Failed to create radvd instance: %v", err)
								}
								fmt.Printf("radvd instance created successfully (%d)\n", id)
							}(c, radvdConfig.Rule.ID, string(file))
							// if err := client.Create(radvdConfig.Rule.ID, string(file)); err != nil {
							// 	log.Fatalf("Failed to create radvd instance: %v", err)
							// }
							// fmt.Printf("radvd instance created successfully (%d)\n", radvdConfig.Rule.ID)
						}
					}
				}
			}
		}
		wg.Wait()
		fmt.Println("radvd instances created successfully (all)")
	case "GET":
		// not implemented

	case "DELETE":
		// not implemented

	default:
		log.Fatalf("Unknown or unsupported method: %s. Use POST, GET, or DELETE", *methodFlag)
	}
}
