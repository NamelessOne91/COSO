build:
	@go build -o ./bin/coso

run: build
	./bin/coso

test:
	go test ./...