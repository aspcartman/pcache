vendor: Gopkg.toml
	dep ensure

run: vendor
	go run ./cmd/main.go

build/pcache: vendor $(find . -path ./vendor -prune -o -name '*.go')
	GOOS=linux GOARCH=amd64 go build -i -o ./build/pcache ./cmd/main.go

build/img: build/pcache
	docker-compose build
	touch ./build/img

clean:
	rm -rf ./build
	rm -rf ./vendor
	rm -rf ./data

test:
	go test ./...

.PHONY: run clean 
