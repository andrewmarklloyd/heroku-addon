.PHONY: build test

GIT_REV=`git rev-parse --short HEAD`
GIT_TREE_STATE=$(shell (git status --porcelain | grep -q .) && echo $(GIT_REV)-dirty || echo $(GIT_REV))

build:
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o bin/heroku-addon *.go

build-frontend:
	cd frontend; npm install; REACT_APP_STRIPE_PUBLIC_KEY=$(REACT_APP_STRIPE_PUBLIC_KEY) npm run build

build-ci: build build-frontend
	cp ./bin/* .

vet:
	go vet ./...

test:
	go test ./...

clean:
	rm -rf bin/
	rm -rf frontend/build/
