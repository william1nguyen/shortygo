build:
	go build -o bin/main cmd/shortygo/main.go

run:
	go run cmd/shortygo/main.go

test:
	go test ./...

clean:
	rm -rf bin/