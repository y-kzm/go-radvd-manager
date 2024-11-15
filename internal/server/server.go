/**
 * radvd manager server.
 * Run the server and handle the REST API requests on IPv6 router.
 * API design see: docs/API.md
 *
 */
package server

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/y-kzm/go-radvd-manager/internal/config"
	"github.com/y-kzm/go-radvd-manager/internal/radvd"
)

type RadvdManagerServer struct {
	http.Server
	radvd  *radvd.Radvd
	logger *slog.Logger
}

func NewServer(host string, logger *slog.Logger) (*RadvdManagerServer, error) {
	existing, err := config.ParseRadvdConfigs()
	if err != nil {
		return nil, fmt.Errorf("failed to parse existing radvd configurations: %v", err)
	}
	srv := &RadvdManagerServer{
		radvd:  existing,
		logger: logger,
	}

	router := mux.NewRouter()
	router.HandleFunc("/restconf/data/radvd:interfaces", srv.handleRadvdInterfaces).Methods("GET", "DELETE")
	router.HandleFunc("/restconf/data/radvd:interfaces/{instance}", srv.handleRadvdInterfaces).Methods("GET", "POST", "PUT", "DELETE")

	srv.Addr = host
	srv.Handler = router

	return srv, nil
}

// for debugging
func (s *RadvdManagerServer) printInterfaces() {
	for _, iface := range s.radvd.Interfaces {
		jsonData, err := json.MarshalIndent(iface, "", "  ")
		if err != nil {
			log.Printf("Error marshaling interface to JSON: %v", err)
			continue
		}
		fmt.Printf("%s", jsonData)
	}
	fmt.Println()
}

func getInterfaceByInstance(interfaces []*radvd.Interface, instance uint32) (*radvd.Interface, error) {
	for _, iface := range interfaces {
		if iface.Instance == instance {
			return iface, nil
		}
	}
	return nil, fmt.Errorf("interface with instance %d not found", instance)
}

func (s *RadvdManagerServer) handleRadvdInterfaces(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceStr, ok := vars["instance"]
	instance, err := strconv.Atoi(instanceStr)
	if (err != nil && instanceStr != "") || (instanceStr != "" && instance <= 0) {
		s.logger.Error("Invalid Instance ID", "instance", instanceStr)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		s.logger.Info("[GET]", "from", r.RemoteAddr)
		w.Header().Set("Content-Type", "application/json")
		if !ok {
			s.logger.Info("Returning all interfaces")
			if err := json.NewEncoder(w).Encode(s.radvd.Interfaces); err != nil {
				s.logger.Error("Failed to encode JSON", "error", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}
		iface, err := getInterfaceByInstance(s.radvd.Interfaces, uint32(instance))
		if err != nil {
			s.logger.Error("Failed to get interface", "error", err.Error())
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err := json.NewEncoder(w).Encode(iface); err != nil {
			s.logger.Error("Failed to encode JSON", "error", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		//s.printInterfaces()
		return
	case "POST":
		s.logger.Info("[POST]", "from", r.RemoteAddr)
		var iface radvd.Interface
		if _, err := getInterfaceByInstance(s.radvd.Interfaces, uint32(instance)); err == nil {
			s.logger.Error("Interface already exists", "instance", instance)
			w.WriteHeader(http.StatusConflict)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&iface); err != nil {
			s.logger.Error("Failed to decode JSON", "error", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if iface.Instance != uint32(instance) {
			s.logger.Error("Instance ID mismatch", "instance", instance, "instance in body", iface.Instance)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		/* generate radvd config */
		if err := config.GenerateRadvdConfigFile(&iface); err != nil {
			s.logger.Error("Failed to generate radvd config", "error", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		/* start radvd */
		if err := radvd.CheckRadvdConfig(int(iface.Instance)); err != nil {
			s.logger.Error("Failed to check radvd config", "error", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err = radvd.StartRadvd(int(iface.Instance)); err != nil {
			s.logger.Error("Failed to start radvd", "error", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		s.radvd.Interfaces = append(s.radvd.Interfaces, &iface)
		w.WriteHeader(http.StatusCreated)
		//s.printInterfaces()
		return
	case "PUT":
		s.logger.Info("[PUT]", "from", r.RemoteAddr)
		var iface radvd.Interface
		existing, err := getInterfaceByInstance(s.radvd.Interfaces, uint32(instance))
		if err != nil {
			s.logger.Error("Interface not found", "instance", instance)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&iface); err != nil {
			s.logger.Error("Failed to decode JSON", "error", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if iface.Instance != uint32(instance) {
			s.logger.Error("Instance ID mismatch", "instance", instance, "instance in body", iface.Instance)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		/* update radvd config */
		if err := config.GenerateRadvdConfigFile(&iface); err != nil {
			s.logger.Error("Failed to generate radvd config", "error", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		/* restart radvd */
		if err := radvd.CheckRadvdConfig(int(iface.Instance)); err != nil {
			s.logger.Error("Failed to check radvd config", "error", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err = radvd.ReloadRadvd(int(iface.Instance)); err != nil {
			s.logger.Error("Failed to reload radvd", "error", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		*existing = iface
		w.WriteHeader(http.StatusCreated)
		//s.printInterfaces()
		return
	case "DELETE":
		s.logger.Info("[DELETE]", "from", r.RemoteAddr)
		if !ok {
			s.logger.Info("Deleting all interfaces")
			/* stop all radvd and delete all radvd config */
			for _, iface := range s.radvd.Interfaces {
				if iface.Instance == 0 {
					continue
				}
				if err := radvd.StopRadvd(int(iface.Instance)); err != nil {
					s.logger.Error("Failed to stop radvd", "error", err.Error())
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
			s.radvd.Interfaces = []*radvd.Interface{}
			w.WriteHeader(http.StatusNoContent)
			//s.printInterfaces()
			return
		}
		existing, err := getInterfaceByInstance(s.radvd.Interfaces, uint32(instance))
		if err != nil {
			s.logger.Error("Interface not found", "instance", instance)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		/* stop radvd adn delete radvd config */
		if err = radvd.StopRadvd(int(existing.Instance)); err != nil {
			s.logger.Error("Failed to stop radvd", "error", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		for i, iface := range s.radvd.Interfaces {
			if iface == existing {
				s.radvd.Interfaces = append(s.radvd.Interfaces[:i], s.radvd.Interfaces[i+1:]...)
				break
			}
		}
		w.WriteHeader(http.StatusNoContent)
		//s.printInterfaces()
		return
	default:
		s.logger.Error("Method not allowed")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func (s *RadvdManagerServer) CleanUp() error {
	for _, iface := range s.radvd.Interfaces {
		if iface.Instance == 0 {
			continue
		}
		radvd.StopRadvd(int(iface.Instance))
	}

	files, err := filepath.Glob("/etc/radvd.d/*")
	if err != nil {
		s.logger.Error("Failed to glob files in /etc/radvd.d/", "error", err.Error())
	} else {
		for _, file := range files {
			if err := os.Remove(file); err != nil {
				s.logger.Error("Failed to remove config file", "file", file, "error", err.Error())
			}
		}
	}

	pidFiles, err := filepath.Glob("/var/run/radvd/radvd.*")
	if err != nil {
		s.logger.Error("Failed to glob PID files in /var/run/radvd/", "error", err.Error())
	} else {
		for _, pidFile := range pidFiles {
			if err := os.Remove(pidFile); err != nil {
				s.logger.Error("Failed to remove PID file", "file", pidFile, "error", err.Error())
			}
		}
	}

	return nil
}
