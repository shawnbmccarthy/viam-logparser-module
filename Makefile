logparser: cmd/module/cmd.go
	@mkdir -p ./bin
	GOOS=android CGO_ENABLED=0 GOARCH=arm64 go build -o ./bin/viam-logparser-module cmd/module/cmd.go

client: cmd/client/cmd.go
	@mkdir -p ./bin
	GOARCH=arm64 go build -o ./bin/lpclient cmd/client/cmd.go

test:
	go test

lint:
	gofmt -w -s .

rm-test-log-dirs:
	@echo "remove existing log testing directories"
	@rm -rf ./tests/logs
	@rm -rf ./tests/upload

mk-log-dirs: rm-test-log-dirs
	@mkdir -p ./tests/logs
	@mkdir -p ./tests/upload

remote-test: mk-log-dirs
	@echo "remote-test"
	./tests/build_logs.sh

clean: rm-test-log-dirs
