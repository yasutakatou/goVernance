/*
 * Governance implementation tools in a push multi-account environment implemented in the Go language
 *
 * @author    yasutakatou
 * @copyright 2025 yasutakatou
 * @license   MIT license, Apache License, version 2.0
 */
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/andreyvit/diff"
	"github.com/aws/aws-lambda-go/lambda"
)

var (
	parallel, lambdamode, debug, logging, noexceptions bool
	replacestr, shell, path                            string
	forcealert, input, whitelist, blacklist            []string
	defines                                            []defineStruct
)

type defineStruct struct {
	Name    string
	Command []string
	Limit   int
	Alert   string
}

type MyEvent struct {
	Name string `json:"name"`
}

func main() {
	_Debug := flag.Bool("debug", false, "[-debug=debug mode (true is enable)]")
	_Logging := flag.Bool("log", false, "[-log=logging mode (true is enable)]")
	_Config := flag.String("config", "governance.ini", "[-config=config file)]")
	_Define := flag.String("define", "define.ini", "[-define=define file)]")
	_NoExceptions := flag.Bool("noexceptions", false, "[-noexceptions=Do not allow everything that is not on the whitelist (true is enable)]")
	_Tmpath := flag.String("tmppath", "/tmp/", "[-tmppath=Output temporary path of the source file to be compared]")
	_Shell := flag.String("shell", "/bin/bash", "[-shell=Specifies the shell to use in the case of linux]")
	_Replacestr := flag.String("replacestr", "{}", "[-replacestr=Replacement string used for forced alerts.]")
	_Parallel := flag.Bool("parallel", false, "[-parallel=Mode to execute tasks in parallel (true is enable)]")

	flag.Parse()

	debug = bool(*_Debug)
	logging = bool(*_Logging)
	noexceptions = bool(*_NoExceptions)
	parallel = bool(*_Parallel)
	shell = string(*_Shell)
	path = string(*_Tmpath)
	replacestr = string(*_Replacestr)

	if os.Getenv("LAMBDA") == "on" {
		lambdamode = true
		shell = os.Getenv("SHELL")
		path = os.Getenv("TMPPATH")
		replacestr = os.Getenv("REPLACESTR")
	} else {
		lambdamode = false
	}

	if os.Getenv("DEBUG") == "on" {
		debug = true
	}
	if os.Getenv("LOG") == "on" {
		logging = true
	}
	if os.Getenv("PARALLEL") == "on" {
		parallel = true
	}
	if os.Getenv("NOEXCEPTIONS") == "on" {
		noexceptions = true
	}

	if os.Getenv("LAMBDA") == "on" {
		debugLog("-- Load Config (" + os.Getenv("CONFIG") + ") --")
		loadConfig(os.Getenv("CONFIG"))
	} else {
		debugLog("-- Load Config (" + *_Config + ") --")
		loadConfig(*_Config)
	}

	if os.Getenv("LAMBDA") == "on" {
		debugLog("-- Define Get (" + os.Getenv("DEFINE") + ") --")
		defineGet(os.Getenv("DEFINE"))
	} else {
		debugLog("-- Define Get (" + *_Define + ") --")
		defineGet(*_Define)
	}

	if os.Getenv("LAMBDA") == "on" {
		debugLog("-- Load Define (" + os.Getenv("DEFINE") + ") --")
		loadDefine(os.Getenv("DEFINE"))
	} else {
		debugLog("-- Load Define (" + *_Define + ") --")
		loadDefine(*_Define)
	}

	debugLog("-- Run Command --")
	if lambdamode == true {
		debugLog("lambda mode: on")
		lambda.Start(HandleRequest)
	} else {
		debugLog("lambda mode: off")
		if parallel == true {
			wg := new(sync.WaitGroup)

			for i := 0; i < len(defines); i++ {
				wg.Add(1)
				go func(n int) {
					checkResult(defines[i])
					wg.Done()
				}(i)
			}
			wg.Wait()
		} else {
			for i := 0; i < len(defines); i++ {
				checkResult(defines[i])
			}
		}
	}
	os.Exit(0)
}

func HandleRequest(ctx context.Context) (*string, error) {
	if parallel == true {
		wg := new(sync.WaitGroup)

		for i := 0; i < len(defines); i++ {
			wg.Add(1)
			go func(n int) {
				checkResult(defines[i])
				wg.Done()
			}(i)
		}
		wg.Wait()
	} else {
		for i := 0; i < len(defines); i++ {
			checkResult(defines[i])
		}
	}

	message := fmt.Sprintf("governance done!")
	return &message, nil
}

func checkResult(define defineStruct) {
	filename := strings.Replace(define.Name, " ", "_", -1)
	filename = strings.Replace(filename, "　", "_", -1)
	filename = "." + filename

	before := ReadFile(path + filename)
	if before == "" {
		debugLog("no exits before result: " + path + filename)
		debugLog("action: " + define.Name)
		after, flag := cmdExecs(define)
		if flag == true {
			Writefile(path+filename, after)
		}
	} else {
		after, flag := cmdExecs(define)
		if flag == true {
			diffs := diff.LineDiff(after, before)
			debugLog(" -- diff -- ")
			debugLog(diffs)
			debugLog(" -- -- -- ")
			cntDiff := countDiff(diffs)
			if cntDiff > define.Limit {
				if len(forcealert) > 0 {
					for r := 0; r < len(forcealert); r++ {
						tmpStr := strings.Replace(forcealert[r], replacestr, define.Name, -1)
						debugLog("[Force] Alert: " + tmpStr)
						cmdExec("", tmpStr)
					}
				} else {
					debugLog("Alert: " + define.Alert)
					cmdExec("", define.Alert)
				}
				Writefile(path+filename, after)
			} else {
				debugLog("No Alert")
				Writefile(path+filename, after)
			}

		}
	}
}

func countDiff(diffs string) int {
	cnt := 0

	for _, v := range regexp.MustCompile("\r\n|\n\r|\n|\r").Split(diffs, -1) {
		regex := regexp.MustCompile(`^[+].*`)
		if regex.MatchString(v) == true {
			cnt = cnt + 1
		}
		regex = regexp.MustCompile(`^[-].*`)
		if regex.MatchString(v) == true {
			cnt = cnt + 1
		}
	}
	return cnt
}

func cmdExecs(define defineStruct) (string, bool) {
	for i := 0; i < len(define.Command)-1; i++ {
		_, flag := cmdExec(define.Name, define.Command[i])
		if flag == false {
			return "", false
		}
	}
	return cmdExec(define.Name, define.Command[len(define.Command)-1])
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

func cmdExec(name, command string) (string, bool) {
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

	debugLog("action: [" + name + "] command: " + command)

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", command)

	} else {
		cmd = exec.Command(shell, "-c", command)
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
		strs, flag := cmdExec("", input[i])
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
		if len(str) > 1 {
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
	forcealert = configRead(filename, "forcealert")
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
