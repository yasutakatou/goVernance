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
