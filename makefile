BUILDTIME=$(shell date '+%Y-%m-%d %H:%M:%S')
BUILDUSER=$(shell whoami)
BUILDHOST=$(shell hostname -f)
BUILDVERSION=$(shell cat ./version.ini)
BUILDBRANCH=$(shell git rev-parse --abbrev-ref HEAD)
BUILDCOMMIT=$(shell git rev-parse --short HEAD)
BUILDOS=$(shell go env GOOS)
BUILDARCH=$(shell go env GOARCH)

PROJECT=codeup.aliyun.com/66d825f8c06a2fdac7bbfe8c/eagleeye

go build  -ldflags="-w -s -X 'codeup.aliyun.com/66d825f8c06a2fdac7bbfe8c/curescan/server/core/meta.BuildVer=$(cat version.ini)'" -o curescan .

LDFLAGS=-w -s -X '${PROJECT}/internal/meta.BuildOS=${BUILDOS}' -X '${PROJECT}/internal/meta.BuildArch=${BUILDARCH}' -X '${PROJECT}/internal/meta.BuildCommit=${BUILDCOMMIT}' -X '${PROJECT}/internal/meta.BuildBranch=${BUILDBRANCH}' -X '${PROJECT}/internal/meta.BuildVer=${BUILDVERSION}' -X '${PROJECT}/internal/meta.BuildTime=${BUILDTIME}' -X '${PROJECT}/internal/meta.BuildUser=${BUILDUSER}' -X '${PROJECT}/internal/meta.BuildHost=${BUILDHOST}'

DOCKER=$(shell which docker)

.PHONY: build linux-docker linux-remote windows darwin swagger clean

build:
	@CGO_ENABLED=1 go build -ldflags "${LDFLAGS}" ./cmd/eagleeye.go

linux-docker:
ifndef DOCKER
	$(error "docker is not available, please install docker")
endif
	@docker build -t eagleeye-builder .
	@docker create --name eagleeye-container eagleeye-builder
	@docker cp eagleeye-container:/src/eagleeye ./eagleeye
	@docker rm eagleeye-container

linux-remote:
	@./build.sh -p ${P} -h ${H} -u ${U}

windows:
	@GOOS=windows CGO_ENABLED=1 go build -ldflags "${LDFLAGS}" ./cmd/eagleeye.go

darwin:
	@GOOS=darwin CGO_ENABLED=1 go build -ldflags "${LDFLAGS}" ./cmd/eagleeye.go

swagger:
	@go generate ./docs/gen.go

clean:
	@rm -rf ./*.csv
	@rm -rf ./*.xlsx
	@rm -rf ./*.out
	@rm -rf ./*.txt
	@rm -rf ./eagleeye*
	@rm -rf ./*.png
	@rm -rf ./results
	@rm -rf ./license.json
	@rm -rf ./*.docx
	@rm -rf ./reports