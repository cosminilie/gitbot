# A Self-Documenting Makefile: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html

PACKAGE := github.com/cosminilie/gitbot
PACKAGECMD :=  github.com/cosminilie/gitbot/cmd
COMMIT_HASH := `git rev-parse --short HEAD 2>/dev/null`
BUILD_DATE := `date -u +%Y%m%d%H%M%S`
GIT_MAJOR_VERSION := `git describe --abbrev=0 --tags| cut  -d'.' -f1`
GIT_MINOR_VERSION := `git describe --abbrev=0 --tags| cut  -d'.' -f2-3``
LDFLAGS := -ldflags "-X main.majorVersion=${GIT_MAJOR_VERSION} -X main.minorVersion=${GIT_MINOR_VERSION} -X main.gitVersion=${COMMIT_HASH} -X main.buildDate=${BUILD_DATE}"
NOGI_LDFLAGS := -ldflags "-X main.buildDate=${BUILD_DATE}"

.PHONY: vendor docker check fmt lint test test-race vet test-cover-html help
.DEFAULT_GOAL := help

vendor: ## Install govendor and sync vendored dependencies
	go get github.com/kardianos/govendor
	govendor sync ${PACKAGE}

gitbot: vendor ## Build gitbot binary
	go build ${LDFLAGS} ${PACKAGECMD}

gitbot-race: vendor ## Build gitbot binary with race detector enabled
	go build -race ${LDFLAGS} ${PACKAGECMD}

install: vendor ## Install gitbot binary
	go install ${LDFLAGS} ${PACKAGECMD}

gitbot-no-gitinfo: LDFLAGS = ${NOGI_LDFLAGS}
gitbot-no-gitinfo: vendor gitbot ## Build gitbot without git info

docker: ## Build gitbot Docker container
	docker build -t gitbot-build . 
	docker rm -f gitbot-build || true
	docker run --name gitbot gitbot-build ls /go/bin
	docker cp gitbot:/go/bin/gitbot .
	docker rm gitbot

docker-rpm: ## Build rpm from docker image
	docker build -t gitbot-rpm -f build/Dockerfile .
	docker rm -f gitbotrpm || true
	docker run --name gitbotrpm gitbot-rpm /bin/bash -c "cd /builds/cosminilie/gitbot; make rpm"
	docker cp gitbotrpm:/root/rpmbuild/RPMS/x86_64 .
	docker rm gitbotrpm

rpm: export GOPATH = /tmp/gopath
rpm: export CI_PROJECT_DIR = /builds/cosminilie/gitbot
rpm: ## Build RPM
	rpmdev-setuptree
	yum-builddep gitbot.spec -y
	spectool -g -R gitbot.spec
	rpmbuild -ba gitbot.spec

check: vendor test-race fmt vet ## Run tests and linters

test: ## Run tests
	govendor test +local

test-race: ## Run tests with race detector
	govendor test -race +local

fmt: ## Run gofmt linter
	@for d in `govendor list -no-status +local | sed -e 's|github.com/cosminilie/gitbot|.|'` ; do \
		if [ "`gofmt -l $$d/*.go | tee /dev/stderr`" ]; then \
			echo "^ improperly formatted go files" && echo && exit 1; \
		fi \
	done

lint: ## Run golint linter
	@for d in `govendor list -no-status +local |  sed -e 's|github.com/cosminilie/gitbot|.|'`` ; do \
		if [ "`golint $$d | tee /dev/stderr`" ]; then \
			echo "^ golint errors!" && echo && exit 1; \
		fi \
	done

vet: ## Run go vet linter
	@if [ "`govendor vet +local | tee /dev/stderr`" ]; then \
		echo "^ go vet errors!" && echo && exit 1; \
	fi

test-cover-html: PACKAGES = $(shell govendor list -no-status +local |  sed -e 's|github.com/cosminilie/gitbot|.|)
test-cover-html: ## Generate test coverage report
	echo "mode: count" > coverage-all.out
	$(foreach pkg,$(PACKAGES),\
		govendor test -coverprofile=coverage.out -covermode=count $(pkg);\
		tail -n +2 coverage.out >> coverage-all.out;)
	go tool cover -html=coverage-all.out

help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
