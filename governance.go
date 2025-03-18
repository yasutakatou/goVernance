package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	debug, logging               bool
	define, whitelist, blacklist []string
	defines                      []defineStruct
)

type defineStruct struct {
	Name    string
	Command string
	Limit   int
	Alert   string
}

func main() {
	_Debug := flag.Bool("debug", false, "[-debug=debug mode (true is enable)]")
	_Logging := flag.Bool("log", false, "[-log=logging mode (true is enable)]")
	_Config := flag.String("config", "governance.ini", "[-config=config file)]")
	_Define := flag.String("define", "define.ini", "[-define=define file)]")

	flag.Parse()

	debug = bool(*_Debug)
	logging = bool(*_Logging)

	loadConfig(*_Config)

	loadDefine(*_Define)

	// define run

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

func configRead(filename, sectionName string) []string {
	var strs []string
	rFlag := false

	debugLog(" -- [" + sectionName + "] --")
	data, _ := os.Open(filename)
	defer data.Close()

	scanner := bufio.NewScanner(data)
	for scanner.Scan() {
		str := scanner.Text()
		if len(str) > 0 {
			if rFlag == true && str[0] == 91 {
				break
			} else {
				if rFlag == true {
					debugLog(str)
					strs = append(strs, str)
				}
			}

			if "["+sectionName+"]" == str {
				rFlag = true
			}
		}
	}
	return strs
}

func loadConfig(filename string) {
	define = configRead(filename, "define")
	whitelist = configRead(filename, "whitelist")
	blacklist = configRead(filename, "blacklist")
}

func loadDefine(filename string) {
	defines = configRead(filename, "define")
}
