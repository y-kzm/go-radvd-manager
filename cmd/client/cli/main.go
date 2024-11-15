/**
 * @file main.go
 * @brief Main function for the client CLI
 */
package main

import (
	"flag"
	"fmt"
	"log"
	"sync"

	"github.com/y-kzm/go-radvd-manager/internal/client"
	"github.com/y-kzm/go-radvd-manager/internal/config"
	"github.com/y-kzm/go-radvd-manager/internal/radvd"
)

const (
	port = 8888
)

type radvdInterfaceAlias = radvd.Interface

func main() {
	execFlag := flag.String("x", "", "Mode: get, apply, update, delete all")
	fileFlag := flag.String("f", "", "Policy config file")
	flag.Parse()

	// read policy config file
	cfg, err := config.LoadPolicyConfig(*fileFlag)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// print config file content
	fmt.Println("----- Policies: -----")
	for _, policy := range cfg.Policies {
		fmt.Printf("ID: %d, Type: %s, FQDN: %s, Prefix: %s, Nexthop: %s\n",
			policy.ID, policy.Type, policy.FQDNs, policy.Prefixes, policy.Nexthop)
	}
	fmt.Println("----- Groups: -----")
	for _, group := range cfg.Groups {
		fmt.Printf("ID: %d, Description: %s, Rules: %v, Clients: %v\n",
			group.ID, group.Description, group.Policies, group.Members)
	}

	fmt.Println("----- Radvd: -----")
	radvd, err := config.ConfigToRadvd(cfg)
	if err != nil {
		log.Fatalf("Failed to convert config to radvd: %v", err)
	}
	fmt.Println("created radvd struct successfully: see ./output")

	// for debug
	// radvdJSON, err := json.MarshalIndent(radvd, "", "  ")
	// if err != nil {
	// 	log.Fatalf("Failed to marshal radvd to JSON: %v", err)
	// }
	// fmt.Println(string(radvdJSON))

	if *fileFlag == "" || *execFlag == "" {
		log.Fatal("Usage: -x <get|apply|update|delete> -f <policy.yaml>")
		return
	}

	// create client
	var clients []*client.RadvdManagerClient
	servers := client.GetSeverList(radvd)

	for _, server := range servers {
		client := client.NewClient(fmt.Sprintf("http://[%s]:%d", server, port), server, port)
		clients = append(clients, client)
	}

	var wg sync.WaitGroup

	// send request
	switch *execFlag {
	case "get":
		for _, c := range clients {
			wg.Add(1)
			go func(client *client.RadvdManagerClient) {
				defer wg.Done()
				radvd, err := client.GetAllInstance()
				if err != nil {
					log.Fatalf("Failed to get radvd instance: %v", err)
				}
				fmt.Println("radvd instance retrieved successfully")
				// for debug
				for _, iface := range radvd.Interfaces {
					fmt.Printf("%+v\n", iface)
				}
			}(c)
		}
		wg.Wait()
	case "apply":
		for _, c := range clients {
			for _, iface := range radvd.Interfaces {
				if iface.Nexthop == c.Server {
					wg.Add(1)
					go func(client *client.RadvdManagerClient, iface *radvdInterfaceAlias) {
						defer wg.Done()
						if err := client.CreateInstance(int(iface.Instance), iface); err != nil {
							log.Fatalf("Failed to apply radvd instance: %v", err)
						}
						fmt.Println("radvd instance applied successfully")
					}(c, iface)
				}
				wg.Wait()
			}
		}
	case "update":
		for _, c := range clients {
			for _, iface := range radvd.Interfaces {
				if iface.Nexthop == c.Server {
					wg.Add(1)
					go func(client *client.RadvdManagerClient, iface *radvdInterfaceAlias) {
						defer wg.Done()
						if err := client.UpdateInstance(int(iface.Instance), iface); err != nil {
							log.Fatalf("Failed to update radvd instance: %v", err)
						}
						fmt.Println("radvd instance updated successfully")
					}(c, iface)
				}
				wg.Wait()
			}
		}
	case "delete":
		for _, c := range clients {
			wg.Add(1)
			go func(client *client.RadvdManagerClient) {
				defer wg.Done()
				err := client.DeleteAllInstance()
				if err != nil {
					log.Fatalf("Failed to delete radvd instances: %v", err)
				}
				fmt.Println("radvd instance deleted successfully")
			}(c)
		}
		wg.Wait()

	default:
		log.Fatalf("Unknown or unsupported method: %s. Use POST, GET, or DELETE", *execFlag)
	}
}
