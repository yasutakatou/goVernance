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

This makes it possible to catch up on changes in usage without having to access the resource usage of each cloud!

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

- Run this app on the account for which you want to implement controls
	- 

# config

This tool uses two types of configurations

## Configuration for operation setting (default: governance.ini)

Definitions are set tab-separated

### [input]

Write a command to get the configuration for the action (see below)
<br>
note) Multiple commands can be listed. The result of the second and subsequent commands will be added to the action configuration in the form of a postscript.

### [ouput]

This command outputs the results of the execution. The replacement string (default: {}) contains the alert message defined in the configuration for the action
<br>
note1) Multiple outputs can be listed. They are executed in order from top to bottom
<br>
note2) The definition before the command specifies the name of the configuration definition for the action. In other words, it is possible to change the notification destination depending on the action

### [whitelist]

Sets the string of commands that can be executed. This prevents unauthorized commands from being defined and executed in the action configuration.
<br>

### [blacklist]

Sets a string of commands that are not allowed to be executed. This prevents unauthorized commands from being defined and executed in the action configuration.
<br>
note) They are evaluated in the order of whitelist to blacklist. In other words, even if you try to filter on the black list, if it is defined on the whitelist, it will be executed.
<br>

```
[input]
C:\Windows\System32\curl.exe http://localhost:3000/def.ini

[output]
file list up	echo "{}"
file list up	aws sns publish --topic-arn {SNS_TOPIC_ARN} --message "{}" --subject "{}"

[whitelist]
dir /b

[blacklist]
dir
```
## Cconfig for action (default: define.ini)

Definitions are set tab-separated

### [define]

Define the action name, difference threshold, and alert content to be notified if the threshold is exceeded

### [(actions)]

Define the name of the action in [ ] and the command to execute. Multiple commands can be listed. The last command executed is compared to the last execution, and if the difference exceeds the threshold defined in [define], an alert is issued.
<br>
note) Diff comparisons are made for both increased and decreased values, similar to the Linux diff command. That is, if one resource is replaced by another name, it is treated as two differences.

```
[define]
file list up	1	file list diff
s3 list command	1	s3 add > 1

[file list up]
dir /b

[s3 list command]
aws s3 ls
```