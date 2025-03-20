package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	debug, logging                      bool
	input, output, whitelist, blacklist []string
	defines                             []defineStruct
)

type defineStruct struct {
	Name    string
	Command []string
	Limit   int
	Alert   string
}

// FYI: https://journal.lampetty.net/entry/capturing-stdout-in-golang
type Capturer struct {
	saved         *os.File
	bufferChannel chan string
	out           *os.File
	in            *os.File
}

func main() {
	_Debug := flag.Bool("debug", false, "[-debug=debug mode (true is enable)]")
	_Logging := flag.Bool("log", false, "[-log=logging mode (true is enable)]")
	_Config := flag.String("config", "governance.ini", "[-config=config file)]")
	_Define := flag.String("define", "define.ini", "[-define=define file)]")
	_NoExceptions := flag.Bool("noexceptions", false, "[-noexceptions=Do not allow all but the whitelist (true is enable)]")

	flag.Parse()

	debug = bool(*_Debug)
	logging = bool(*_Logging)

	loadConfig(*_Config)
	defineGet(*_Define)
	loadDefine(*_Define)
	runCommand(*_NoExceptions)

	os.Exit(0)
}

func runCommand(noexceptions bool) {
	for i := 0; i < len(defines); i++ {
		if checkWhitelist(defines[i].Command) == true {
			checkResult(defines[i].Command)
		} else {
			if noexceptions == false && checkBacklist(defines[i].Command) == false {
				checkResult(defines[i].Command)
			}
		}
	}
}

func checkResult(command string) {

}

func checkWhitelist(command string) bool {
	for i := 0; i < len(whitelist); i++ {
		if strings.Index(command, whitelist[i]) != -1 {
			return true
		}
	}
	return false
}

func checkBlacklist(command string) bool {
	for i := 0; i < len(blacklist); i++ {
		if strings.Index(command, blacklist[i]) != -1 {
			return true
		}
	}
	return false
}

func cmdExec(command string) string {
	var cmd *exec.Cmd

	debugLog("command: " + command)

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", command)

	} else {
		cmd = exec.Command(command)
	}

	output, err := cmd.Output()
	if err != nil {
		fmt.Println("コマンド実行エラー:", err)
		return ""
	}

	debugLog(string(output))
	return string(output)
}

// 標準出力をキャプチャする
func (c *Capturer) StartCapturingStdout() {
	c.saved = os.Stdout
	var err error
	c.in, c.out, err = os.Pipe()
	if err != nil {
		panic(err)
	}

	os.Stdout = c.out
	c.bufferChannel = make(chan string)
	go func() {
		var b bytes.Buffer
		io.Copy(&b, c.in)
		c.bufferChannel <- b.String()
	}()
}

// キャプチャを停止する
func (c *Capturer) StopCapturingStdout() string {
	c.out.Close()
	os.Stdout = c.saved
	return <-c.bufferChannel
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
		fmt.Fprint(file, cmdExec(input[i]))
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
				if rFlag == true && str[0] != 35 {
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
	output = configRead(filename, "output")
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
