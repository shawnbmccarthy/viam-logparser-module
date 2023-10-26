logparser: cmd/module/cmd.go
	GOOS=android CGO_ENABLED=0 GOARCH=arm64 go build -o viam-logparser-module cmd/module/cmd.go

test:
	go test

lint:
	gofmt -w -s .