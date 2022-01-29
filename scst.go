package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

const SCST_ROOT_PATH string = "/sys/kernel/scst_tgt"
const SCST_DEVICES string = SCST_ROOT_PATH + "/devices"
const SCST_ISCSI_TARGETS string = SCST_ROOT_PATH + "/targets/iscsi"
const SYSFS_SCST_DEV_MGMT string = SCST_ROOT_PATH + "/handlers/vdisk_blockio/mgmt"
const SYSFS_SCST_LUNS_MGMT string = "/ini_groups/allowed_ini/luns/mgmt"

type ScstLun struct {
	Filename string `json:"filename"`
	DevId    string `json:"devid"`
}

type ScstResponse struct {
	Error    error             `json:"error"`
	DataList []string          `json:"datalist"`
	DataMap  map[string]string `json:"datamap"`
}

func readParamsFromDir(dirpath string) (map[string]string, error) {
	var (
		res map[string]string
		err error
		val []byte
	)
	res = make(map[string]string)
	if dataDir, err := os.Open(dirpath); err != nil {
		return res, err
	} else {
		if files, err := dataDir.ReadDir(0); err != nil {
			return res, err
		} else {
			for _, v := range files {
				if f, err := os.Open(path.Join(dirpath, v.Name())); err != nil {
					res[v.Name()] = err.Error()
				} else {
					if val, err = ioutil.ReadAll(f); err != nil {
						res[v.Name()] = err.Error()
					}
					res[v.Name()] = strings.TrimSuffix(string(val), "\n")
					res[v.Name()] = strings.TrimSuffix(res[v.Name()], "\n[key]")
					res[v.Name()] = strings.TrimSuffix(res[v.Name()], "\n[key]\n")
					res[v.Name()] = strings.TrimSuffix(res[v.Name()], "\n")
				}
			}
		}
	}
	return res, err
}

func readFromDir(dirpath string) ([]string, error) {
	var (
		res []string
		err error
	)

	if dataDir, err := os.Open(dirpath); err != nil {
		return res, err
	} else {
		if files, err := dataDir.ReadDir(0); err != nil {
			return res, err
		} else {
			for _, v := range files {
				res = append(res, v.Name())
			}
		}
	}
	return res, err
}

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

func ScstGetIscsiTargets() ScstResponse {
	var (
		res ScstResponse
		err error
	)
	if res.DataList, err = readFromDir(SCST_ISCSI_TARGETS); err != nil {
		log.Println(err.Error())
		res.Error = err
	}
	return res
}

func ScstGetDeviceParam(device string, param string) string {
	var (
		paramPath string
		res       []string
	)

	paramPath = SCST_DEVICES + "/" + device + "/" + param
	deviceParam, err := os.Open(paramPath)
	if err != nil {
		res = append(res, "")
	} else {
		deviceParamData, err := ioutil.ReadAll(deviceParam)
		if err != nil {
			res = append(res, "")
		}
		res = strings.Split(string(deviceParamData), "\n")
	}
	return res[0]

}

func ScstGetIscsiTargetParam(target string, param string) string {
	var (
		paramPath string
		targets   []string
		tgt       string
		res       []string
	)
	targets = ScstGetIscsiTargets().DataList
	for _, v := range targets {
		if strings.Contains(v, target) {
			tgt = v
			break
		}
	}
	if tgt != "" {
		paramPath = SCST_ISCSI_TARGETS + "/" + tgt + "/" + param
		deviceParam, err := os.Open(paramPath)
		if err != nil {
			res = append(res, "")
		} else {
			deviceParamData, err := ioutil.ReadAll(deviceParam)
			if err != nil {
				res = append(res, "")
			}
			res = strings.Split(string(deviceParamData), "\n")
		}
	} else {
		res = append(res, "Target not found!")
	}

	return res[0]

}

func scstSetDeviceParam(device string, param string, val string) {
	var (
		paramPath string
	)

	paramPath = SCST_DEVICES + "/" + device + "/" + param
	if deviceParam, err := os.OpenFile(paramPath, os.O_WRONLY, 0644); err != nil {
		log.Println(err)
	} else {
		defer deviceParam.Close()
		deviceParamW, err := deviceParam.Write([]byte(val))
		if err != nil {
			log.Println(err)
		}
		log.Printf("Wrote %d bytes.\n", deviceParamW)
	}

}

func ScstGetDeviceParams(device string) ScstResponse {
	var (
		res ScstResponse
		err error
	)
	res.DataMap = make(map[string]string)
	if res.DataMap, err = readParamsFromDir(path.Join(SCST_DEVICES, device)); err != nil {
		res.Error = err
	}
	return res
}

func ScstGetIscsiTargetParams(target string) ScstResponse {
	var (
		res     ScstResponse
		targets []string
		wwn     string
		err     error
	)
	res.DataMap = make(map[string]string)
	targets = ScstGetIscsiTargets().DataList
	for _, v := range targets {
		if strings.Contains(v, target) {
			wwn = v
			break
		}
	}
	if wwn != "" {
		if res.DataMap, err = readParamsFromDir(path.Join(SCST_ISCSI_TARGETS, wwn)); err != nil {
			res.Error = err
		}
	}
	return res
}

func ScstDeleteDevice(device string) map[string]string {
	var (
		res map[string]string
	)
	res = make(map[string]string)
	scstCmd := []byte("del_device " + device)
	if mgmt, err := os.OpenFile(SYSFS_SCST_DEV_MGMT, os.O_WRONLY, 0644); err != nil {
		res["error"] = "Cannot open SCST device management interface!"
		log.Println(err.Error())
	} else {
		defer mgmt.Close()
		if mgmtW, err := mgmt.Write(scstCmd); err != nil {
			res["error"] = err.Error()
			log.Println(res["error"])
		} else {
			log.Printf("Wrote %d bytes.\n", mgmtW)
		}
	}
	return res
}

func ScstCreateLun(lun ScstLun) map[string]string {
	var (
		res         map[string]string
		lunPathMgmt string
		wwn         string
	)
	res = make(map[string]string)
	targets := ScstGetIscsiTargets().DataList
	for _, v := range targets {
		if strings.Contains(v, lun.DevId) {
			wwn = v
			break
		}
	}
	if wwn != "" {
		if mgmt, err := os.OpenFile(SYSFS_SCST_DEV_MGMT, os.O_WRONLY, 0644); err != nil {
			res["error"] = "Cannot open SCST device management interface!"
			log.Println(res["error"])
		} else {
			defer mgmt.Close()
			scstCmd := "add_device " + lun.DevId + " filename=" + lun.Filename + "; nv_cache=1; rotational=0"
			if mgmtW, err := mgmt.Write([]byte(scstCmd)); err != nil {
				res["error"] = err.Error()
				log.Println(res["error"])
			} else {
				log.Printf("Wrote %d bytes.\n", mgmtW)
				res["device"] = fmt.Sprintf("Wrote %d bytes.\n", mgmtW)
				lunPathMgmt = SCST_ISCSI_TARGETS + "/" + wwn + SYSFS_SCST_LUNS_MGMT
				if mgmt, err = os.OpenFile(lunPathMgmt, os.O_WRONLY, 0644); err != nil {
					res["error"] = err.Error()
					log.Println(res["error"])
				} else {
					defer mgmt.Close()
					scstCmd = "add " + lun.DevId + " 0"
					if mgmtW, err := mgmt.Write([]byte(scstCmd)); err != nil {
						res["error"] = err.Error()
						log.Println(res["error"])
					} else {
						log.Printf("Wrote %d bytes.\n", mgmtW)
						res["target"] = fmt.Sprintf("Wrote %d bytes.\n", mgmtW)
					}
				}
			}
		}
		return res
	} else {
		res["error"] = fmt.Sprintf("Target %s not found!", lun.DevId)
		log.Println(res["error"])
		return res
	}
}

func ScstListIscsiSessions(target string) ScstResponse {
	var (
		res          ScstResponse
		sessionsPath string
		wwn          string
		err          error
	)
	targets := ScstGetIscsiTargets().DataList
	for _, v := range targets {
		if strings.Contains(v, target) {
			wwn = v
			break
		}
	}
	if wwn != "" {
		sessionsPath = path.Join(SCST_ISCSI_TARGETS, wwn, "sessions")
		if res.DataList, err = readFromDir(sessionsPath); err != nil {
			res.Error = err
		}
	}
	return res
}
