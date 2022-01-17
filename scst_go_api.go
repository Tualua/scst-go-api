package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Tualua/scst_go_api/scst_go"
)

func returnDevices(w http.ResponseWriter, r *http.Request) {
	devices := scst_go.ScstGetDevices()
	fmt.Println("Endpoint Hit: returnDevices")
	json.NewEncoder(w).Encode(devices)
}

func returnIscsiTargets(w http.ResponseWriter, r *http.Request) {
	iscsiTargets := scst_go.ScstGetIscsiTargets()
	fmt.Println("Endpoint Hit: returnIscsiTargets")
	json.NewEncoder(w).Encode(iscsiTargets)
}

func handleRequests() {
	http.HandleFunc("/devices", returnDevices)
	http.HandleFunc("/targets", returnIscsiTargets)
	log.Fatal(http.ListenAndServe(":10000", nil))
}

func main() {
	handleRequests()
}
