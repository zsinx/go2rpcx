version := $(shell /bin/date "+%Y-%m-%d %H:%M")

build:
	go build -ldflags="-s -w" -ldflags="-X 'main.BuildTime=$(version)'" -o go2rpcx.exe main.go
	$(if $(shell command -v upx), upx go2rpcx)
mac:
	GOOS=darwin go build -ldflags="-s -w" -ldflags="-X 'main.BuildTime=$(version)'" -o go2rpcx-darwin main.go
	$(if $(shell command -v upx), upx go2rpcx-darwin)
win:
	GOOS=windows go build -ldflags="-s -w" -ldflags="-X 'main.BuildTime=$(version)'" -o go2rpcx.exe main.go
	$(if $(shell command -v upx), upx go2rpcx.exe)
linux:
	GOOS=linux go build -ldflags="-s -w" -ldflags="-X 'main.BuildTime=$(version)'" -o go2rpcx-linux main.go
	$(if $(shell command -v upx), upx go2rpcx-linux)
