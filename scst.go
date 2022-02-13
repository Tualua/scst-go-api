package main

import (
	"errors"
	"fmt"
	"io/fs"
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

func ScstGetDevices() ([]string, error) {
	var (
		res         []string
		scstDevices []fs.DirEntry
		err         error
	)
	if scstDevicesDir, err := os.Open(SCST_DEVICES); err != nil {
		log.Println(err.Error())
	} else {
		if scstDevices, err = scstDevicesDir.ReadDir(0); err != nil {
			log.Println(err.Error())
		} else {
			for _, v := range scstDevices {
				res = append(res, v.Name())
			}
		}
	}
	return res, err
}

func ScstGetIscsiTargets() (res []string, err error) {
	if res, err = readFromDir(SCST_ISCSI_TARGETS); err != nil {
		log.Println(err.Error())
	}
	return res, err
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

func ScstGetIscsiTargetParam(target string, param string) (res string, err error) {
	var (
		paramPath string
		targets   []string
		tgt       string
		// res       []string
	)
	if targets, err = ScstGetIscsiTargets(); err != nil {
		log.Println(err.Error())
	} else {
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
				res = ""
			} else {
				deviceParamData, err := ioutil.ReadAll(deviceParam)
				if err != nil {
					res = ""
				}
				res = strings.Split(string(deviceParamData), "\n")[0]
			}
		} else {
			res = "Target not found!"
		}
	}

	return
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
	targets, err = ScstGetIscsiTargets()
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

func ScstDeleteDevice(device string) (err error) {
	var (
	// res map[string]string
	)
	// res = make(map[string]string)
	scstCmd := []byte("del_device " + device)
	if mgmt, err := os.OpenFile(SYSFS_SCST_DEV_MGMT, os.O_WRONLY, 0644); err != nil {
		log.Println(err.Error())
	} else {
		defer mgmt.Close()
		if _, err := mgmt.Write(scstCmd); err != nil {
			log.Println(err.Error())
		} else {
			log.Printf("Device %s deleted \n", device)
		}
	}
	return
}

func ScstDeactivateDevice(device string) (err error) {
	var (
		mgmt *os.File
	)
	scstCmd := []byte("0")
	scstCmdPath := path.Join(SCST_DEVICES, device, "active")
	if mgmt, err = os.OpenFile(scstCmdPath, os.O_WRONLY, 0644); err != nil {
		log.Println(err.Error())
	} else {
		defer mgmt.Close()
		if _, err = mgmt.Write(scstCmd); err != nil {
			log.Println(err.Error())
		} else {
			log.Printf("Device %s deactivated \n", device)
		}
	}
	return
}

func ScstActivateDevice(device string) (err error) {
	var (
		mgmt *os.File
	)
	scstCmd := []byte("1")
	scstCmdPath := path.Join(SCST_DEVICES, device, "active")
	if mgmt, err = os.OpenFile(scstCmdPath, os.O_WRONLY, 0644); err != nil {
		log.Println(err.Error())
	} else {
		defer mgmt.Close()
		if _, err = mgmt.Write(scstCmd); err != nil {
			log.Println(err.Error())
		} else {
			log.Printf("Device %s activated \n", device)
		}
	}
	return
}

func ScstCreateLun(devId string, fileName string) error {
	var (
		lunPathMgmt string
		wwn         string
		targets     []string
		err         error
	)
	if targets, err = ScstGetIscsiTargets(); err != nil {
		log.Println(err.Error())
	} else {
		for _, v := range targets {
			if strings.Contains(v, devId) {
				wwn = v
				break
			}
		}
		if wwn != "" {
			if mgmt, err := os.OpenFile(SYSFS_SCST_DEV_MGMT, os.O_WRONLY, 0644); err != nil {
				err = errors.New("Cannot open SCST device management interface!")
				log.Println(err.Error())
			} else {
				defer mgmt.Close()
				scstCmd := "add_device " + devId + " filename=" + fileName + "; nv_cache=1; rotational=0"
				if _, err := mgmt.Write([]byte(scstCmd)); err != nil {
					log.Println(err.Error())
				} else {
					log.Printf("Device %s backed by %s added\n", devId, fileName)
					lunPathMgmt = SCST_ISCSI_TARGETS + "/" + wwn + SYSFS_SCST_LUNS_MGMT
					if mgmt, err = os.OpenFile(lunPathMgmt, os.O_WRONLY, 0644); err != nil {
						log.Println(err.Error())
					} else {
						defer mgmt.Close()
						scstCmd = "add " + devId + " 0"
						if _, err := mgmt.Write([]byte(scstCmd)); err != nil {
							log.Println(err.Error())
						} else {
							log.Printf("LUN 0 with device %s exported via target %s\n", devId, wwn)
						}
					}
				}
			}
		} else {
			err = errors.New(fmt.Sprintf("Target %s not found!", devId))
			log.Println(err.Error())
		}
	}
	return err
}

func ScstListIscsiSessions(target string) (res []string, err error) {
	var (
		sessionsPath string
		wwn          string
		targets      []string
	)
	targets, err = ScstGetIscsiTargets()
	for _, v := range targets {
		s := strings.Split(v, ":")
		if s[len(s)-1] == target {
			wwn = v
			break
		}
	}
	if wwn != "" {
		sessionsPath = path.Join(SCST_ISCSI_TARGETS, wwn, "sessions")
		if res, err = readFromDir(sessionsPath); err != nil {
			log.Println(err.Error())
		}
	} else {
		err = errors.New(fmt.Sprintf("Target with ID %s not found!", target))
	}
	return
}
