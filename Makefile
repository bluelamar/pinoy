
GO_VERSION?=1.11.5
BUILD_NUMBER?=local
GIT_COMMIT?=$(shell git log --pretty=format:'%H' -n 1)
VERSION=$(shell cat ./VERSION)
SVCBIN?=$(GOPATH)/bin/pinoy

PKG_DATE=$(shell date '+%Y-%m-%dT%H:%M:%S')

#BUILD_SERVERS=./pinoy/...
BUILD_SERVERS=./...
BUILD_TOOLS =\
	     ./tools/aws-iam-audit \
	     ./tools/biz-metrics-email \
	     ./tools/cve-notification-email \
	     ./tools/dbaudit \
	     ./tools/devdb \
	     ./tools/weeklydigest/weekly-digest-email

all: buildall

buildall: lint
	# go install ${BUILD_SERVERS} ${BUILD_TOOLS}
	go install -ldflags "-X main.svcVersionString=$(VERSION) -X main.buildNum=$(BUILD_NUMBER) -X main.buildDate=$(PKG_DATE) -X main.buildSHA=$(GIT_COMMIT)" ${BUILD_SERVERS}
	# compile integration test into binary
	#go test -c ./testing/integration/...
	#cp integration.test ${GOPATH}/bin

#start: buildall
#	./scripts/make-start.sh

#start-nobuild:
#	./scripts/make-start.sh

test: buildall unit

unit:
	#go test -short -tags unit $(BUILD_TESTS)
	#go test $(go list ./... | grep -v vendor | grep -v rulelang)
	go test ./...

# let go check for potential issues
vet:
	go vet ./...

fmt:
	gofmt -l *.go config/*go database/*go food/*go misc/*go pemail/*go psession/*go room/*go staff/*go

coverage:
	COMPREHENSIVE_MODE=true ./scripts/make-coverage.sh

pkg-coverage:
	COMPREHENSIVE_MODE=false ./scripts/make-coverage.sh

# create app with race detection tool - depends on local libc or glibc
# alpine = /lib/ld-musl-x86_64.so.1
build-race:
	go build -o pinoy-race -ldflags "-X main.svcVersionString=$(VERSION)+race -X main.buildNum=$(BUILD_NUMBER) -X main.buildSHA=$(GIT_COMMIT)" -race ./pinoy/cmd/pinoy/...

.PHONY: lint
lint:
	echo "TODO no lint"
#	./scripts/make-lint.sh

clean:
	go clean ./...
	rm -f pinoy pinoy-race .tmp
	rm -f coverage.html coverage_fuc.txt coverage.xml profile.cov pkg-coverage.txt


