package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/Tualua/scst_go_api/scst_go"
	"github.com/go-yaml/yaml"
	"github.com/gorilla/mux"
)

type Config struct {
	Server struct {
		Port string `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"server"`
}

func NewConfig(configPath string) (*Config, error) {
	config := &Config{}
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	d := yaml.NewDecoder(file)
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}

func returnDevices(w http.ResponseWriter, r *http.Request) {
	devices := scst_go.ScstGetDevices()
	fmt.Println("Endpoint Hit: returnDevices")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	enc.Encode(devices)
}

func returnIscsiTargets(w http.ResponseWriter, r *http.Request) {
	iscsiTargets := scst_go.ScstGetIscsiTargets()
	fmt.Println("Endpoint Hit: returnIscsiTargets")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	enc.Encode(iscsiTargets)
}

func returnDeviceParams(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	params := scst_go.ScstGetDeviceParams(vars["id"])
	fmt.Printf("Endpoint Hit: returnDeviceParams, Target: %s\n", vars["id"])
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	enc.Encode(params)

}
func returnIscsiTargetParams(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	params := scst_go.ScstGetIscsiTargetParams(vars["id"])
	fmt.Printf("Endpoint Hit: returnTargetParams, Target: %s\n", vars["id"])
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	enc.Encode(params)

}

func returnIscsiTargetSessions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	params := scst_go.ScstListIscsiSessions(vars["id"])
	fmt.Printf("Endpoint Hit: returnTargetParams, Target: %s\n", vars["id"])
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	enc.Encode(params)

}

func deleteDevice(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: deleteDevice")
	reqBody, _ := ioutil.ReadAll(r.Body)
	var device scst_go.ScstLun
	json.Unmarshal(reqBody, &device)
	fmt.Printf("Operation: deleteDevice Device %s", device.DevId)
	res := scst_go.ScstDeleteDevice(device.DevId)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	enc.Encode(res)
}

func createLun(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: createLun")
	reqBody, _ := ioutil.ReadAll(r.Body)
	var device scst_go.ScstLun
	json.Unmarshal(reqBody, &device)
	fmt.Printf("Operation: createLun Device %s", device.DevId)
	res := scst_go.ScstCreateLun(device)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	enc.Encode(res)
}

func handleRequests(cfg *Config) {
	router := mux.NewRouter().StrictSlash(true)
	addrString := cfg.Server.Host + ":" + cfg.Server.Port
	router.HandleFunc("/devices", returnDevices)
	router.HandleFunc("/lun/create", createLun).Methods("POST")
	router.HandleFunc("/device/delete", deleteDevice).Methods("POST")
	router.HandleFunc("/device/{id}", returnDeviceParams)
	router.HandleFunc("/targets", returnIscsiTargets)
	router.HandleFunc("/target/{id}", returnIscsiTargetParams)
	router.HandleFunc("/target/{id}/sessions", returnIscsiTargetSessions)
	log.Fatal(http.ListenAndServe(addrString, router))
}

func main() {
	cfg, err := NewConfig("config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	handleRequests(cfg)
}
