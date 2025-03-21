package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/andreyvit/diff"
)

var (
	debug, logging, noexceptions bool
	input, whitelist, blacklist  []string
	defines                      []defineStruct
)

type defineStruct struct {
	Name    string
	Command []string
	Limit   int
	Alert   string
}

func main() {
	_Debug := flag.Bool("debug", false, "[-debug=debug mode (true is enable)]")
	_Logging := flag.Bool("log", false, "[-log=logging mode (true is enable)]")
	_Config := flag.String("config", "governance.ini", "[-config=config file)]")
	_Define := flag.String("define", "define.ini", "[-define=define file)]")
	_NoExceptions := flag.Bool("noexceptions", false, "[-noexceptions=Do not allow everything that is not on the whitelist (true is enable)]")
	_ReplaceStr := flag.String("replacestr", "{}", "[-replacestr=string to replace in output)]")
	_Path := flag.String("path", "/tmp/", "[-path=Output path of the source file to be compared)]")

	flag.Parse()

	debug = bool(*_Debug)
	logging = bool(*_Logging)
	noexceptions = bool(*_NoExceptions)

	debugLog("-- Load Config --")
	loadConfig(*_Config)
	debugLog("-- Define Get --")
	defineGet(*_Define)
	debugLog("-- Load Define --")
	loadDefine(*_Define)
	debugLog("-- Run Command --")
	for i := 0; i < len(defines); i++ {
		checkResult(defines[i].Command, *_ReplaceStr, *_Path)
	}
	os.Exit(0)
}

func checkResult(command []string, replacestr, path string) {
	for i := 0; i < len(defines); i++ {
		filename := strings.Replace(defines[i].Name, " ", "_", -1)
		filename = strings.Replace(filename, "　", "_", -1)
		filename = "." + filename

		before := ReadFile(path + filename)
		if before == "" {
			debugLog("no exits before result: " + path + filename)
			after, flag := cmdExecs(defines[i].Command)
			if flag == true {
				Writefile(path+filename, after)
			}
		} else {
			after, flag := cmdExecs(defines[i].Command)
			if flag == true {
				diffs := diff.LineDiff(after, before)
				cntDiff := countDiff(diffs)
				if cntDiff > defines[i].Limit {
					debugLog("Alert: " + defines[i].Alert)
					cmdExec(defines[i].Alert)
					Writefile(path+filename, after)
				} else {
					debugLog("No Alert")
				}

			}
		}
	}
}

func countDiff(diffs string) int {
	cnt := 0
	cnt = cnt + strings.Count(diffs, "+")
	cnt = cnt + strings.Count(diffs, "-")
	return cnt
}

func cmdExecs(commands []string) (string, bool) {
	for i := 0; i < len(commands)-1; i++ {
		_, flag := cmdExec(commands[i])
		if flag == false {
			return "", false
		}
	}
	return cmdExec(commands[len(commands)-1])
}

func checkWhitelist(command string) bool {
	for i := 0; i < len(whitelist); i++ {
		regex := regexp.MustCompile(whitelist[i])
		if regex.MatchString(command) == true {
			return true
		}
	}
	return false
}

func checkBlacklist(command string) bool {
	for i := 0; i < len(blacklist); i++ {
		regex := regexp.MustCompile(blacklist[i])
		if regex.MatchString(command) == true {
			return true
		}
	}
	return false
}

func cmdExec(command string) (string, bool) {
	if checkWhitelist(command) == false {
		if noexceptions == true {
			debugLog("no permission whitelist[noexceptions]: " + command)
			return "", false
		}
		if checkBlacklist(command) == true {
			debugLog("no permission blacklist: " + command)
			return "", false
		}
	}

	var cmd *exec.Cmd

	debugLog("command: " + command)

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", command)

	} else {
		cmd = exec.Command(command)
	}

	output, err := cmd.Output()
	if err != nil {
		fmt.Println("コマンド実行エラー:", command, err)
		return "", false
	}

	debugLog(string(output))
	return string(output), true
}

func defineGet(filename string) {
	if len(input) == 0 {
		return
	}

	os.Remove(filename)

	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	for i := 0; i < len(input); i++ {
		strs, flag := cmdExec(input[i])
		if flag == true {
			fmt.Fprint(file, strs)
		}
	}
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

func ReadFile(fileName string) string {
	bytes, err := os.ReadFile(fileName)
	if err != nil {
		return ""
	}

	return string(bytes)
}

func Writefile(filename, strs string) {
	os.Remove(filename)
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	fmt.Fprint(file, strs)
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
			if rFlag == true && str[0] == 91 { // [
				break
			} else {
				if rFlag == true && str[0] != 35 { // #
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
	input = configRead(filename, "input")
	whitelist = configRead(filename, "whitelist")
	blacklist = configRead(filename, "blacklist")
}

func loadDefine(filename string) {
	strs := configRead(filename, "define")

	for i := 0; i < len(strs); i++ {
		splitStr := strings.Split(strs[i], "\t")
		if len(splitStr) == 3 {
			commands := configRead(filename, splitStr[0])
			convInt, err := strconv.Atoi(splitStr[1])
			if err == nil {
				defines = append(defines, defineStruct{Name: splitStr[0], Command: commands, Limit: convInt, Alert: splitStr[2]})
			}
		}
	}
}
