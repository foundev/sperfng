.PHONY: all
all: clean lint build test

.PHONY: clean
clean:
	go run hack/clean.go ./bin/

outfile=sperfng
ifeq ($(OS), Windows_NT)
	outfile=sperfng.exe
endif

.PHONY: build
build:
	go build -o ./bin/$(outfile) .

.PHONY: test
test:
	go test ./...

.PHONY: lint
lint:
	go vet ./...
	go fmt -x ./...