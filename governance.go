package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"gopkg.in/ini.v1"
)

var (
	debug, logging                        bool
	define, whitelist, blacklist, defines []string
)

func main() {
	_Debug := flag.Bool("debug", false, "[-debug=debug mode (true is enable)]")
	_Logging := flag.Bool("log", false, "[-log=logging mode (true is enable)]")
	_Config := flag.String("config", "governance.ini", "[-config=config file)]")
	_Define := flag.String("define", "define.ini", "[-define=define file)]")

	flag.Parse()

	debug = bool(*_Debug)
	logging = bool(*_Logging)

	loadConfig(*_Config)

	// define run

	loadDefine(*_Define)

	os.Exit(0)
}

func debugLog(message string) {
	var file *os.File
	var err error

	if debug == true {
		fmt.Println(message)
	}

	if logging == false {
		return
	}

	const layout = "2006-01-02_15"
	const layout2 = "2006/01/02 15:04:05"
	t := time.Now()
	filename := t.Format(layout) + ".log"
	logHead := "[" + t.Format(layout2) + "] "

	if Exists(filename) == true {
		file, err = os.OpenFile(filename, os.O_WRONLY|os.O_APPEND, 0666)
	} else {
		file, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
	}

	if err != nil {
		log.Fatal(err)
		return
	}
	defer file.Close()
	fmt.Fprintln(file, logHead+message)
}

func Exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func urlget(url string) string {

	resp, _ := http.Get(url)
	defer resp.Body.Close()
	byteArray, _ := io.ReadAll(resp.Body)
	debugLog(string(byteArray))
	return string(byteArray)
}

func loadConfig(configFile string) {
	loadOptions := ini.LoadOptions{}
	loadOptions.UnparseableSections = []string{"define", "whitelist", "blacklist"}

	cfg, err := ini.LoadSources(loadOptions, configFile)
	if err != nil {
		fmt.Printf("Fail to read config file: %v", err)
		os.Exit(1)
	}

	define = setStructs("define", cfg.Section("define").Body())
	whitelist = setStructs("whitelist", cfg.Section("whitelist").Body())
	blacklist = setStructs("blacklist", cfg.Section("blacklist").Body())
}

func setStructs(configType, datas string) []string {
	var strs []string
	debugLog(" -- " + configType + " --")

	for _, v := range regexp.MustCompile("\r\n|\n\r|\n|\r").Split(datas, -1) {
		if len(v) > 0 {
			strs = append(strs, v)
			debugLog(v)
		}
	}
	return strs
}

func loadDefine(defineFile string) {
	//loadOptions := ini.LoadOptions{}
	//loadOptions.UnparseableSections = []string{"define"}

	cfg, err := ini.Load(defineFile)
	if err != nil {
		fmt.Printf("Fail to read config file: %v", err)
		os.Exit(1)
	}

	defines = setStructs("define", cfg.Section("define").Body())
}
