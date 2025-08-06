APP_NAME=lamina

run:
	go run main.go

build:
	go build -o bin/$(APP_NAME) main.go

test:
	go test ./...

fmt:
	go fmt ./...

lint:
	golangci-lint run

clean:
	rm -rf bin/

