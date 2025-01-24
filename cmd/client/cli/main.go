package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	radvd "github.com/y-kzm/go-radvd-manager"
	client "github.com/y-kzm/go-radvd-manager/cmd/internal"
)

const (
	port = 12345
)

func main() {
	execFlag := flag.String("x", "", "[status|apply|update|delete]")
	fileFlag := flag.String("f", "", "Policy file")
	flag.Parse()

	if *execFlag == "" {
		log.Fatalf("Use -x [status|apply|update|delete]")
	}
	policy, err := radvd.LoadPolicyFile(*fileFlag)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	show_policy(policy)
	instances, err := radvd.ParsePolicy(policy)
	if err != nil {
		log.Fatalf("Failed to convert policy to radvd instance: %v", err)
	}
	// create clients
	routers := client.GetSiteExitRouters(instances)
	clients := make([]*client.RadvdManagerClient, len(routers))
	for i, r := range routers {
		client := client.NewClient(fmt.Sprintf("http://[%s]:%d", r, port), r, port)
		clients[i] = client
	}

	// send requests
	switch *execFlag {
	case "status":
		for _, c := range clients {
			if err = c.GetInstances(); err != nil {
				log.Fatalf("Failed to get radvd instances: %v", err)
			}
		}
		show_status(clients)
	case "apply":
		var clientWg sync.WaitGroup
		for _, c := range clients {
			// gorutine for each client
			clientWg.Add(1)
			go func(c *client.RadvdManagerClient) {
				defer clientWg.Done()
				var instanceWg sync.WaitGroup
				for _, i := range instances {
					if i.RouterID == c.Server {
						// gorutine for each instance in the client
						instanceWg.Add(1)
						go func(instance *radvd.Instance) {
							defer instanceWg.Done()
							err := c.CreateInstance(int(i.ID), instance)
							if err != nil {
								log.Printf("Failed to create radvd instance: %v", err)
							} else {
								log.Printf("+ Created radvd instance (id: %d) on %s", i.ID, c.Server)
							}
						}(i)
					}
				}
				instanceWg.Wait()
			}(c)
		}
		clientWg.Wait()
		time.Sleep(5 * time.Second)
	case "update":
		break
	case "delete":
		for _, c := range clients {
			err := c.DeleteInstances()
			if err != nil {
				log.Fatalf("Failed to delete radvd instances: %v", err)
			}
		}
	default:
		log.Fatalf("Unknown or unsupported method: %s. Use POST, GET, or DELETE", *execFlag)
	}
}

func show_policy(policy *radvd.Policy) {
	fmt.Println("[Local Policy]")
	fmt.Printf("%-12s %-40s %-20s\n", "ID(common)", "Prefixes", "Nexthop")
	fmt.Println(strings.Repeat("-", 80))
	for _, i := range policy.Rules {
		prefixes := "[" + strings.Join(i.Prefixes, " ") + "]"
		fmt.Printf("%-12d %-40s %-20s\n", i.ID, prefixes, i.Nexthop)
	}
	fmt.Println("\nRules                Members")
	fmt.Println(strings.Repeat("-", 80))
	for _, i := range policy.Groups {
		rules := strings.Join(strings.Fields(fmt.Sprint(i.Rules)), " ")
		members := "[" + strings.Join(i.Members, " ") + "]"
		fmt.Printf("%-20s %-30s\n", rules, members)
	}
	fmt.Println()
}

func show_status(clients []*client.RadvdManagerClient) {
	fmt.Println("[Remote Status]")
	fmt.Printf("%-20s %-12s %-5s %-40s %-12s %-30s\n", "RouterID", "ID(common)", "PID", "Routes", "Preference", "Clients")
	fmt.Println(strings.Repeat("-", 150))
	for _, c := range clients {
		for _, i := range c.RemoteInstances {
			members := "[" + strings.Join(i.Clients, " ") + "]"
			var routes []string
			for _, r := range i.Routes {
				routes = append(routes, r.Route)
			}
			routes_formated := "[" + strings.Join(routes, " ") + "]"
			fmt.Printf("%-20s %-12d %-5d %-40s %-12s %-30s\n", i.RouterID, i.ID, i.PID, routes_formated, i.AdvDefaultPreference, members)
		}
		fmt.Println()
	}
}
