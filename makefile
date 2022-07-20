
.PHONY: go init ui all

all: ui go

go:
	cd cmd/cli && go build
	go build

init:
	cd ./webUI && npm i

ui:
	cd ./webUI && npm run build
