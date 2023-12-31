export GO_EXECUTABLE_PATH := $(shell which go)

fs-setup:
	@mkdir -p /tmp/coso/rootfs
	@tar -C /tmp/coso/rootfs -xzf assets/alpine-minirootfs-3.18.2-x86_64.tar.gz
	@echo "alpine 3.18 x86_64 filesystem decompressed at /tmp/coso/rootfs"

net-setup:
	@echo "building cosonet"
	@cd cmd/cosonet/ && go build -o ../../bin/cosonet
	@echo "moving cosonet to /usr/local/bin"
	@sudo cp bin/cosonet /usr/local/bin
	@echo "chown cosonet to the root user"
	@sudo chown root:root /usr/local/bin/cosonet
	@echo "apply the setuid bit for cosonet"
	@sudo chmod 4755 /usr/local/bin/cosonet
	@echo "Done!"

build:
	@cd cmd/coso/ && go build -o ../../bin/coso

run: build
	@./bin/coso

test:
	@echo $$GO_EXECUTABLE_PATH
	@sudo $$GO_EXECUTABLE_PATH test -v ./...
