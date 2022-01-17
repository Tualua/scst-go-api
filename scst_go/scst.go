package scst_go

import (
	"log"
	"os"
)

const SCST_ROOT_PATH string = "/sys/kernel/scst_tgt"
const SCST_DEVICES string = SCST_ROOT_PATH + "/devices"
const SCST_ISCSI_TARGETS string = SCST_ROOT_PATH + "/targets/iscsi"

func ScstGetDevices() []string {
	var res []string
	scstDevicesDir, err := os.Open(SCST_DEVICES)
	if err != nil {
		if os.IsNotExist(err) {
			log.Fatal("SCST devices directory does not exist!")
		}
	}
	scstDevices, err := scstDevicesDir.ReadDir(0)
	for _, v := range scstDevices {
		res = append(res, v.Name())
	}
	return res
}

func ScstGetIscsiTargets() []string {
	var res []string
	scstIscsiTargetsDir, err := os.Open(SCST_ISCSI_TARGETS)
	if err != nil {
		if os.IsNotExist(err) {
			log.Fatal("SCST iSCSI targets directory does not exist!")
		}
	}
	scstIscsiTargets, err := scstIscsiTargetsDir.ReadDir(0)
	for _, v := range scstIscsiTargets {
		if v.IsDir() {
			res = append(res, v.Name())
		}
	}
	return res
}
