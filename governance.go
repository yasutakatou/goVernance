package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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

	flag.Parse()

	debug = bool(*_Debug)
	logging = bool(*_Logging)

	loadConfig(*_Config)
	defineGet(*_Define)
	loadDefine(*_Define)
	runCommand()

	os.Exit(0)
}

func runCommand() {
	for i := 0; i < len(defines); i++ {
		//if checkWhitelist(defines[i].Command) == true && checkBacklist(defines[i].Command) == true {
		//	result := cmdExec(defines[i].Command)
		//}
	}
}

func cmdExec(command string) string {
	var cmd *exec.Cmd
	var out string

	debugLog("command: " + command)

	cmd = exec.Command(os.Getenv("SHELL"), "-c", command)

	c := &Capturer{}
	c.StartCapturingStdout()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Run()

	out = c.StopCapturingStdout()
	debugLog(out)
	return out
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
		splitStr := strings.Split(input[i], "\t")
		switch splitStr[0] {
		case "s3":
			fmt.Fprint(file, s3get(splitStr[1], splitStr[2]))
		case "url":
			fmt.Fprint(file, urlget(splitStr[1]))
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

func urlget(url string) string {
	resp, _ := http.Get(url)
	defer resp.Body.Close()
	byteArray, _ := io.ReadAll(resp.Body)
	debugLog(string(byteArray))
	return string(byteArray)
}

func s3get(bucket, key string) string {
	// AWS設定の読み込み
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	client := s3.NewFromConfig(cfg)

	output, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		log.Fatalf("failed to get object, %v", err)
	}
	defer output.Body.Close()

	body, err := ioutil.ReadAll(output.Body)
	if err != nil {
		log.Fatalf("failed to read object body, %v", err)
	}

	debugLog(string(body))
	return string(body)
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
