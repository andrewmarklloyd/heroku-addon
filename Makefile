.PHONY: build test

GIT_REV=`git rev-parse --short HEAD`
GIT_TREE_STATE=$(shell (git status --porcelain | grep -q .) && echo $(GIT_REV)-dirty || echo $(GIT_REV))

build:
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o bin/heroku-addon *.go

vet:
	go vet ./...

test:
	go test ./...

clean:
	rm -rf build/
	rm -rf frontend/build/
