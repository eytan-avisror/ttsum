.PHONY: test docker clean all

COMMIT=`git rev-parse HEAD`
BUILD=`date +%FT%T%z`
LDFLAG_LOCATION=github.com/eytan-avisror/ttsum/cmd/cli

LDFLAGS=-ldflags "-X ${LDFLAG_LOCATION}.buildDate=${BUILD} -X ${LDFLAG_LOCATION}.gitCommit=${COMMIT}"

GIT_TAG=$(shell git rev-parse --short HEAD)

build:
	CGO_ENABLED=0 go build ${LDFLAGS} -o bin/ttsum github.com/eytan-avisror/ttsum
	chmod +x bin/ttsum

test:
	go test -v ./... -coverprofile coverage.txt
	go tool cover -html=coverage.txt -o coverage.html
