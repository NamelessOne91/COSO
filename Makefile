fssetup:
	mkdir -p /tmp/coso/rootfs
	tar -C /tmp/coso/rootfs -xvzf assets/alpine-minirootfs-3.18.2-x86_64.tar.gz

build:
	@go build -o ./bin/coso

run: build
	./bin/coso

test:
	go test ./...