package main

import (
	"flag"
	"fmt"
	"log"

	//"github.com/y-kzm/go-radvd-manager/internal/client"
	"github.com/y-kzm/go-radvd-manager/internal/config"
)

const (
	port = 8888
)

func main() {
	//hostFlag := flag.String("h", "localhost", "Hostname")
	//methodFlag := flag.String("X", "", "HTTP Method (POST, GET, DELETE)")
	fileFlag := flag.String("f", "", "Config file")
	idFlag := flag.Int("i", -1, "Instance ID")
	flag.Parse()

	//client := client.NewClient(fmt.Sprintf("http://%s:%d", *hostFlag, port))
	if *idFlag == -1 {
		log.Fatal("Usage: -i <id> is required")
	}

	// 設定ファイルを読み込む
	cfg, err := config.LoadPolicyConfig(*fileFlag)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// パースされたルールとポリシーを表示する
	fmt.Println("Rules:")
	for _, rule := range cfg.Rules {
		fmt.Printf("ID: %d, Type: %s, FQDN: %s, Prefix: %s, Nexthop: %s\n",
			rule.ID, rule.Type, rule.FQDNs, rule.Prefixes, rule.Nexthop)
	}

	fmt.Println("\nPolicies:")
	for _, policy := range cfg.Policies {
		fmt.Printf("ID: %d, Description: %s, Rules: %v, Clients: %v\n",
			policy.ID, policy.Description, policy.Rules, policy.Clients)
	}

	/*

		switch *methodFlag {
		case "POST":
			if *fileFlag == "" {
				log.Fatal("Usage: -X POST -f <config_file> -i <id> -h <hostname>")
			}
			config, err := os.ReadFile(*fileFlag)
			if err != nil {
				log.Fatalf("Failed to read config file: %v", err)
			}
			if err := client.Create(*idFlag, string(config)); err != nil {
				log.Fatalf("Failed to create radvd instance: %v", err)
			}
			fmt.Println("radvd instance created successfully")

		case "GET":
			config, err := client.Get(*idFlag)
			if err != nil {
				log.Fatalf("Failed to get radvd instance: %v", err)
			}
			var data map[string]string
			if err := json.Unmarshal([]byte(config), &data); err != nil {
				log.Fatalf("Failed to unmarshal config: %v", err)
			}
			fmt.Println(data["config"])

		case "DELETE":
			if err := client.Delete(*idFlag); err != nil {
				log.Fatalf("Failed to delete radvd instance: %v", err)
			}
			fmt.Println("radvd instance deleted successfully")

		default:
			log.Fatalf("Unknown or unsupported method: %s. Use POST, GET, or DELETE", *methodFlag)
		}
	*/
}
