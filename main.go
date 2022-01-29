package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func returnDevices(w http.ResponseWriter, r *http.Request) {
	devices := ScstGetDevices()
	fmt.Println("Endpoint Hit: returnDevices")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	enc.Encode(devices)
}

func returnIscsiTargets(w http.ResponseWriter, r *http.Request) {
	iscsiTargets := ScstGetIscsiTargets()
	fmt.Println("Endpoint Hit: returnIscsiTargets")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	enc.Encode(iscsiTargets)
}

func returnDeviceParams(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	params := ScstGetDeviceParams(vars["id"])
	fmt.Printf("Endpoint Hit: returnDeviceParams, Target: %s\n", vars["id"])
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	enc.Encode(params)

}
func returnIscsiTargetParams(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	params := ScstGetIscsiTargetParams(vars["id"])
	fmt.Printf("Endpoint Hit: returnTargetParams, Target: %s\n", vars["id"])
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	enc.Encode(params)

}

func returnIscsiTargetSessions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	params := ScstListIscsiSessions(vars["id"])
	fmt.Printf("Endpoint Hit: returnTargetParams, Target: %s\n", vars["id"])
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	enc.Encode(params)

}

func deleteDevice(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: deleteDevice")
	reqBody, _ := ioutil.ReadAll(r.Body)
	var device ScstLun
	json.Unmarshal(reqBody, &device)
	fmt.Printf("Operation: deleteDevice Device %s", device.DevId)
	res := ScstDeleteDevice(device.DevId)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	enc.Encode(res)
}

func createLun(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: createLun")
	reqBody, _ := ioutil.ReadAll(r.Body)
	var device ScstLun
	json.Unmarshal(reqBody, &device)
	fmt.Printf("Operation: createLun Device %s", device.DevId)
	res := ScstCreateLun(device)
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
