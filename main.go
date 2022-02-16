package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var (
	pathConfigProd = flag.String("config", "/etc/scstapi/config.yaml", "Path to config.yaml")
	pathConfigDev  = "config.yaml"
	pathConfig     string
)

func loggingMiddleware(next http.Handler) http.Handler {
	return handlers.CombinedLoggingHandler(os.Stdout, next)
}

func apiListDevices(w http.ResponseWriter, r *http.Request) {
	var (
		res jsonResponseList
		err error
	)
	res.Action = "devices"
	if res.Data, err = ScstGetDevices(); err != nil {
		res.Error(err.Error())
	} else {
		res.Success()
	}
	res.Write(&w)
}

func apiListIscsiTargets(w http.ResponseWriter, r *http.Request) {
	var (
		iscsiTargets []string
		res          jsonResponseList
		err          error
	)
	res.SetAction("targets")
	if iscsiTargets, err = ScstGetIscsiTargets(); err != nil {
		res.Error(err.Error())
	} else {
		res.Data = iscsiTargets
		res.Success()
	}

	res.Write(&w)
}

func apiGetDeviceParams(w http.ResponseWriter, r *http.Request) {
	var (
		res    jsonResponseMapString
		err    error
		params map[string]string
	)
	res.SetAction("devparams")
	if params, err = ScstGetDeviceParams(mux.Vars(r)["devid"]); err != nil {
		res.Error(err.Error())
	} else {
		res.Success()
		res.Data = params
	}

	res.Write(&w)

}
func apiListIscsiTargetParams(w http.ResponseWriter, r *http.Request) {
	var (
		res    jsonResponseMapString
		params map[string]string = make(map[string]string)
		err    error
	)
	res.SetAction("iscsitargetparams")
	if params, err = ScstGetIscsiTargetParams(mux.Vars(r)["tgtid"]); err != nil {
		res.Error(err.Error())
	} else {
		res.Success()
		res.Data = params
	}
	res.Write(&w)
}

func apiListIscsiSessions(w http.ResponseWriter, r *http.Request) {
	var (
		res jsonResponseList
		err error
	)
	res.SetAction("iscsisessions")
	if res.Data, err = ScstListIscsiSessions(mux.Vars(r)["tgtid"]); err != nil {
		res.Error(err.Error())
	} else {
		res.Success()
	}
	res.Write(&w)
}

func apiDeleteDevice(w http.ResponseWriter, r *http.Request) {
	var (
		res jsonResponseList
		err error
	)
	res.SetAction("deletedev")
	if err = ScstDeleteDevice(mux.Vars(r)["devid"]); err != nil {
		res.Error(err.Error())
	} else {
		res.Success()
	}

	res.Write(&w)
}

func apiActivateDevice(w http.ResponseWriter, r *http.Request) {
	var (
		res jsonResponseList
		err error
	)
	res.SetAction("actdev")
	if err = ScstActivateDevice(mux.Vars(r)["devid"]); err != nil {
		res.Error(err.Error())
	} else {
		res.Success()
	}
	res.Write(&w)
}

func apiDeactivateDevice(w http.ResponseWriter, r *http.Request) {
	var (
		res jsonResponseList
		err error
	)
	res.SetAction("deactdev")
	if err = ScstDeactivateDevice(mux.Vars(r)["devid"]); err != nil {
		res.Error(err.Error())
	} else {
		res.Success()
	}
	res.Write(&w)
}

func apiCreateLun(w http.ResponseWriter, r *http.Request) {
	var (
		err error
		res jsonResponseList
	)
	res.SetAction("createlun")
	if err = ScstCreateLun(mux.Vars(r)["devid"], mux.Vars(r)["filename"]); err != nil {
		res.Error(err.Error())
	} else {
		res.Success()
	}
	res.Write(&w)
}

func run(cfg *Config) {
	router := mux.NewRouter().StrictSlash(true)
	addrString := cfg.Server.Host + ":" + cfg.Server.Port
	router.Path("/").Queries("action", "devices").HandlerFunc(apiListDevices)
	router.Path("/").Queries("action", "createlun", "devid", "{devid}", "filename", "{filename}").HandlerFunc(apiCreateLun)
	router.Path("/").Queries("action", "deletedev", "devid", "{devid}").HandlerFunc(apiDeleteDevice)
	router.Path("/").Queries("action", "actdev", "devid", "{devid}").HandlerFunc(apiActivateDevice)
	router.Path("/").Queries("action", "deactdev", "devid", "{devid}").HandlerFunc(apiDeactivateDevice)
	router.Path("/").Queries("action", "devparams", "devid", "{devid}").HandlerFunc(apiGetDeviceParams)
	router.Path("/").Queries("action", "iscsitargets").HandlerFunc(apiListIscsiTargets)
	router.Path("/").Queries("action", "iscsitargetparams", "tgtid", "{tgtid}").HandlerFunc(apiListIscsiTargetParams)
	router.Path("/").Queries("action", "iscsisessions", "tgtid", "{tgtid}").HandlerFunc(apiListIscsiSessions)
	router.Use(loggingMiddleware)
	log.Fatal(http.ListenAndServe(addrString, router))
}

func main() {
	if os.Getenv("APP_ENV") == "dev" {
		pathConfig = pathConfigDev
		log.Println("Running in development environment")
	} else {
		flag.Parse()
		pathConfig = *pathConfigProd
	}

	if cfg, err := NewConfig(pathConfig); err != nil {
		log.Fatal(err)
	} else {
		log.Printf("Using config file %s", pathConfig)
		if os.Getenv("APP_ENV") != "dev" && cfg.Server.Host != "127.0.0.1" {
			log.Printf("SCST API will be listening on %s. Using other than 127.0.0.1 address is NOT RECOMMENDED for production evironment!", cfg.Server.Host)
		}
		run(cfg)
	}
}
