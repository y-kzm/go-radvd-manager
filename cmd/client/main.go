package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/y-kzm/go-radvd-manager/internal"
)

func main() {
	client := internal.NewClient("http://localhost:8888")

	if len(os.Args) < 2 {
		log.Fatal("Usage: create/get")
	}

	switch os.Args[1] {
	case "create":
		if len(os.Args) < 4 {
			log.Fatal("Usage: create <id> <config>")
		}
		config, err := os.ReadFile(os.Args[3])
		if err != nil {
			log.Fatalf("Failed to read config file: %v", err)
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatalf("Failed to parse id: %v", err)
		}
		if err := client.Create(id, string(config)); err != nil {
			log.Fatalf("Failed to create radvd instance: %v", err)
		}
		fmt.Println("radvd instance created successfully")
	case "get":
		if len(os.Args) < 3 {
			log.Fatal("Usage: get <id>")
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatalf("Failed to parse id: %v", err)
		}
		config, err := client.Get(id)
		if err != nil {
			log.Fatalf("Failed to get radvd instance: %v", err)
		}
		var data map[string]string
		if err := json.Unmarshal([]byte(config), &data); err != nil {
			log.Fatalf("Failed to unmarshal config: %v", err)
		}

		fmt.Println(data["config"])
	case "delete":
		if len(os.Args) < 3 {
			log.Fatal("Usage: delete <id>")
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatalf("Failed to parse id: %v", err)
		}
		if err := client.Delete(id); err != nil {
			log.Fatalf("Failed to delete radvd instance: %v", err)
		}
		fmt.Println("radvd instance deleted successfully")
	default:
		log.Fatal("Unknown command")
	}
}
