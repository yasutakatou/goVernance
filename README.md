# goVernance
**Governance implementation tools in a push multi-account environment implemented in the Go language.**

# Solution
Applying governance to the various accounts that exist within a company is not easy. Moreover, it is very difficult to visualize the security requirements if there is **no access authority to each account**.
If there are no access privileges, we thought that we could apply governance to the accounts themselves by generating information as a batch process and sending the information in a **push type**, while still maintaining security requirements.

# Feature
- Executes commands defined at runtime and updates the definition file
- Outputs the results of the execution to a file and compares them with the results of the previous execution
- Compare the execution results of the defined commands and execute the defined command as an alert if the difference is large

In other words,

- Set up a destination to get the definition file for this code and schedule execution on a serverless service in the cloud
- The results of the execution are compared with the results of the previous execution, and an alert is issued if there is a large difference.
- Add settings, including additional definitions, according to the governance defined by the organization

**This makes it possible to catch up on changes in usage without having to access the resource usage of each cloud!**

# install

If you want to put it under the path, you can use the following.

```
go get github.com/yasutakatou/goVernance
```

If you want to create a binary and copy it yourself, use the following.

```
git clone https://github.com/yasutakatou/goVernance
cd goVernance
go build goVernance.go
```

or download binary from [release page](https://github.com/yasutakatou/goVernance/releases).
save binary file, copy to entryed execute path directory.

# uninstall

```
delete that binary.
```

del or rm command. (it's simple!)

# usage

- Sets where the control rules are obtained. Also, determine the white list and black list and prepare the configuration file
- Upload your app onto a computing resource, such as the cloud
	- EC2, Lambda, or any container is acceptable, but Lambda is recommended because it needs to run on a schedule.
	- For serverless operation, specify a non-volatile area such as /tmp to refer to the output of the previous run
- Attach the permissions you want the app to have access to
	- For example, if you want to control the S3 list, you can add a permission to see the S3 list.
	- It is a good idea to verify the operation in a sandbox environment such as Cloud shell
- Set the information you want to control in the configuration for the action
- With periodic execution, if the difference from the previous operation exceeds a threshold value, the command for the configured alert will be executed and you will be notified.

This allows you to receive alerts when differences occur in resource information, etc., without having to obtain information on the side of the account you want to control, and without having to obtain access rights!<br>

- If you want to add resource information or other information that you want to control, edit the file and add an item to the configuration for the action

You can add the items you want to control by simply appending them to the config for the action! You can also whitelist and blacklist commands that the account administrator does not want to be executed.<br>
<br>
note) **It is recommended to always include a setting to skip alerts and to determine if the notifier is working with the notified party according to the scheduled execution, such as ping, and to monitor the alerts.**

# config

This tool uses two types of configurations

## Configuration for operation setting (default: governance.ini)

### [input]

Write a command to get the configuration for the action (see below)
<br>
note) Multiple commands can be listed. The result of the second and subsequent commands will be **added to the action configuration** in the form of a postscript.

### [whitelist]

Sets the string of commands that can be executed. This prevents unauthorized commands from being defined and executed in the action configuration.
<br>
note) Multiple lines of listing can be specified. Also, the specification must be a regular expression.

### [blacklist]

Sets a string of commands that are not allowed to be executed. This prevents unauthorized commands from being defined and executed in the action configuration.
<br>
note1) They are evaluated in the order of whitelist to blacklist. In other words, **even if you try to filter on the black list, if it is defined on the whitelist, it will be executed**.
<br>
note2) Multiple lines of listing can be specified. Also, the specification must be a regular expression.

### config example

```
[input]
C:\Windows\System32\curl.exe http://localhost:3000/define.ini

[whitelist]
dir /b

[blacklist]
dir
```

## Cconfig for action (default: define.ini)

### [define]

format (Definitions are set tab-separated)

```
(action name)	(threshold)	(alert command)
```

Define the action name, difference threshold, and alert command to be notified if the threshold is exceeded
<br>
note) Diff comparisons are made for both increased and decreased values, **similar to the Linux diff command**. That is, if one resource is replaced by another name, it is treated as two differences.

### [(actions)]

Define the name of the action in [ ] and the command to execute. Multiple commands can be listed. **The last command executed is compared to the last execution**, and if the difference exceeds the threshold defined in [define], an alert is issued.

### config example

```
[define]
file list up	1	echo "file list diff"
s3 list command	1	aws sns publish --topic-arn {SNS_TOPIC_ARN} --message "s3 add > 1" --subject "s3 diff alert"

[file list up]
dir /b

[s3 list command]
aws s3 ls
```

# options

This application has a **normal mode** and **lambda mode**, like running on EC2, and the options are set differently

## normal mode

This is the mode in which options are specified as strings from the **command line at startup**

```
  -config string
        [-config=config file)] (default "governance.ini")
  -debug
        [-debug=debug mode (true is enable)]
  -define string
        [-define=define file)] (default "define.ini")
  -log
        [-log=logging mode (true is enable)]
  -noexceptions
        [-noexceptions=Do not allow everything that is not on the whitelist (true is enable)]
  -path string
        [-path=Output path of the source file to be compared)] (default "/tmp/")
  -replacestr string
        [-replacestr=string to replace in output)] (default "{}")
```

### -config

Specify the path to the configuration file

### -debug

Runs in debug mode. Various output is printed.

### -define

Specify the path of the config for the action

### -log

Option to output the log from debug mode.

### -noexceptions

This mode treats commands that are not on the white list as blacklisted and does not allow them to be executed.<br>
<br>
note) **By default, it operates in a lax mode**, executing commands that are not on the whitelist and not on the blacklist.

### -shell

Specifies the shell to use in the case of linux

### -tmppath

Specify the path where the temporary file from the previous run is to be created. The temporary file will be created with spaces converted to underscores in the configuration definition for the action<br>
<br>
note1) **For serverless operation, specify a non-volatile area such as /tmp to refer to the output of the previous run**<br>
<br>
note2) Note that the path specification changes when running on Windows compared to Linux. ( / -> \\\ )<br>

```
goVernance.exe -path=".\\tmp\\"
```

## lambda mode

This mode is to start as AWS Lambda. In this case, specify the options as **environment variables** for lambda

### LAMBDA

Starts in lambda mode. Set the environment variable **LAMBDA** to **on**

### CONFIG

Specify the path to the configuration file. Specify a **string** in the environment variable

### DEBUG

Runs in debug mode. Various output is printed. Set the environment variable **DEBUG** to **on**

### DEFINE

Specify the path of the config for the action. Specify a **string** in the environment variable

### LOG

Option to output the log from debug mode. Set the environment variable **LOG** to **on**

### NOEXCEPTIONS

This mode treats commands that are not on the white list as blacklisted and does not allow them to be executed. Set the environment variable **NOEXCEPTIONS** to **on**<br>
<br>
note) **By default, it operates in a lax mode**, executing commands that are not on the whitelist and not on the blacklist.

### SHELL

Specifies the shell to use in the case of linux. Specify a **string** in the environment variable

### TMPPATH

Specify the path where the temporary file from the previous run is to be created. The temporary file will be created with spaces converted to underscores in the configuration definition for the action. Specify a **string** in the environment variable<br>
<br>
note1) **For serverless operation, specify a non-volatile area such as /tmp to refer to the output of the previous run**<br>
<br>
note2) Note that the path specification changes when running on Windows compared to Linux. ( / -> \\\ )<br>

# lisence
MIT license
