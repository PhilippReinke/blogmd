GIT_VERSION := $(shell git rev-parse HEAD)

test:
	go test ./...

install:
	go install -ldflags "-X github.com/PhilippReinke/blogmd/args.GitCommit=$(GIT_VERSION)" .

build:
	go build -ldflags "-X github.com/PhilippReinke/blogmd/args.GitCommit=$(GIT_VERSION)"

clean:
	go clean
	go clean -testcache
