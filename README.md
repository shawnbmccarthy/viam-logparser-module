# Android Log Parser

Simple log rotation module for orange pi android devices

## Programs

1. [client](./cmd/client/cmd.go): client code to run sensor doCommand and find log files of a given pattern
2. [module](./cmd/module/cmd.go): module code to start the viam module from the rdk config
3. [remote](./cmd/remote/cmd.go): test server for validating sensor component works

## Build

### compile android module
```shell
make logparser
```

upload viam-logparser-module up to android system using adb commands (testing only)
