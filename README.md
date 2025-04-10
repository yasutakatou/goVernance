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

## lambda mode

Example of operating in **lambda mode**<br>
Create a lambda layer in aws cli

[Lambda関数でAWS CLIを使用する](https://qiita.com/DaichiAndoh/items/c738a0f12c51e30f9b93)

Create a lambda function. The **timeout should be sufficient** for each command to execute.<br>
(If the lambda limit is exceeded, please separate the functions or take other measures.)

![image](https://github.com/user-attachments/assets/c2f3db3b-81a6-4ab5-9742-ea635dd832f0)

Set parameters to environment variables

![image](https://github.com/user-attachments/assets/db0e4131-b327-447e-a7ab-4959fa3ad64d)

Upload the golang zip file to the lambda function. **Include the following configurations**(governance.ini) in the zip file

```
[input]
aws s3 cp s3://defines-s3-backet/define.ini /tmp/define.ini

[whitelist]
.*aws s3 ls.*
.*aws s3 cp s3://defines-s3-backet/define.ini.*
.*aws sns publish --topic-arn "arn:aws:sns:ap-northeast-1:128259705520:email".*

[blacklist]
.*aws ec2.*

[forcealert]
```

On the S3 side, describe the commands to be compared<br>
**Attach to lambda a policy** and its role that allows lambda functions access to S3

```
[define]
s3 list up	1	aws sns publish --topic-arn "arn:aws:sns:ap-northeast-1:128259705520:email" --message "s3 add > 1" --subject "s3 diff alert"

[s3 list up]
aws s3 ls
```

lambda compares the differences at each execution. Create two new S3 buckets as follows

![image](https://github.com/user-attachments/assets/a9942868-298b-4735-a739-5ec7b690d122)

function detects differences and executes an action because of one or more differences

```
Status: Succeeded
Test Event Name: test

Response:
"governance done!"

Function Logs:
START RequestId: 65b1723d-8fe1-4565-a3be-10d666e6b317 Version: $LATEST
command: aws s3 ls
2025-03-23 08:40:02 defines-s3-backet
2025-03-27 06:43:27 defines-s3-backet2
2025-03-27 06:43:35 defines-s3-backet3
-- diff -- 
2025-03-23 08:40:02 defines-s3-backet
-2025-03-27 06:43:27 defines-s3-backet2
-2025-03-27 06:43:35 defines-s3-backet3
-- -- -- 
Alert: aws sns publish --topic-arn "arn:aws:sns:ap-northeast-1:128259705520:email" --message "s3 add > 1" --subject "s3 diff alert"
command: aws sns publish --topic-arn "arn:aws:sns:ap-northeast-1:128259705520:email" --message "s3 add > 1" --subject "s3 diff alert"
{
"MessageId": "2d62a728-c0a5-5764-9304-c77636ae3d95"
}
END RequestId: 65b1723d-8fe1-4565-a3be-10d666e6b317
REPORT RequestId: 65b1723d-8fe1-4565-a3be-10d666e6b317	Duration: 26906.24 ms	Billed Duration: 26907 ms	Memory Size: 128 MB	Max Memory Used: 119 MB

Request ID: 65b1723d-8fe1-4565-a3be-10d666e6b317
```

You have been notified by email according to your SNS action

![image](https://github.com/user-attachments/assets/141f9ae0-5e6e-42f3-a880-0a3d98138bd2)

note) Please continue to consider a visualization mechanism, such as linking email notifications with Power Platform, etc., to create a dashboard as numerical values.

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

### [forcealert] (v0.2-)

**Override the alert notification commands** on the config side if you do not want the config for action to use any of the alert notification commands.<br>
<br>
note) The notification is specified in the **replacement string**(default: {}) with the name of the action in the configuration for the action

### config example

```
[input]
C:\Windows\System32\curl.exe http://localhost:3000/define.ini

[whitelist]
dir /b

[blacklist]
dir

[forcealert]
echo "force alert! : {}"
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
  -parallel
	[-parallel=Mode to execute tasks in parallel (true is enable)]
  -replacestr string
        [-replacestr=string to replace in output)] (default "{}")
  -tmppath string
        [-tmppath=Output path of the source file to be compared)] (default "/tmp/")
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

### -parallel (v0.3-)

Mode to execute tasks in parallel<br>
<br>
note) Parallel execution mode when there are time constraints, such as in lambda execution mode.

### -replacestr

The action name will be replaced by the definition of this string when the alert command is forced<br>
<br>
config example)
```
[forcealert]
echo "force alert! : {}"
```

action config example)

```
[define]
file list up	1	echo "file list diff"
```

output)

```
force alert! : file list up
```

note) This mode applies to all of the [forcealert] sections of the config

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

### PARALLEL (v0.3-)

Mode to execute tasks in parallel<br>
<br>
note) Parallel execution mode when there are time constraints, such as in lambda execution mode.

### REPLACESTR

The action name will be replaced by the definition of this string when the alert command is forced<br>
<br>
config example)
```
[forcealert]
echo "force alert! : {}"
```

action config example)

```
[define]
file list up	1	echo "file list diff"
```

output)

```
force alert! : file list up
```

### SHELL

Specifies the shell to use in the case of linux. Specify a **string** in the environment variable

### TMPPATH

Specify the path where the temporary file from the previous run is to be created. The temporary file will be created with spaces converted to underscores in the configuration definition for the action. Specify a **string** in the environment variable<br>
<br>
note1) **For serverless operation, specify a non-volatile area such as /tmp to refer to the output of the previous run**<br>
<br>
note2) Note that the path specification changes when running on Windows compared to Linux. ( / -> \\\ )<br>

# lisence
MIT license<br>
Apache License, version 2.0
