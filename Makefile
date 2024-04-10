.DEFAULT_GOAL := build

clean:
	rm -f dorset-binformation

fmt:
	go fmt ./...
.PHONY:fmt

lint: fmt
	golint ./...
.PHONY:lint

vet: fmt
	go vet ./...
.PHONY:vet

build: vet
	go build
.PHONY:build

run: build
	./dorset-binformation
.PHONY:run

install: build
	mkdir -p /opt/dorset-binformation/
	useradd dorset-binformation || true
	chown dorset-binformation:dorset-binformation /opt/dorset-binformation/
	cp dorset-binformation /opt/dorset-binformation/
	cp dorset-binformation.service /etc/systemd/system/