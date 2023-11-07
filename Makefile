ifeq ($(OS),Windows_NT)
	HOSTOS=windows
	ifeq ($(PROCESSOR_ARCHITEW6432),AMD64)
		HOSTARCH=amd64
	else
		ifeq ($(PROCESSOR_ARCHITECTURE),AMD64)
			HOSTARCH=amd64
		else ifeq ($(PROCESSOR_ARCHITECTURE),x86)
			HOSTARCH=amd32
		endif
	endif
else
	UNAME_S := $(shell uname -s)
	ifeq ($(UNAME_S),Linux)
		HOSTOS=linux
	else ifeq ($(UNAME_S),Darwin)
		HOSTOS=darwin
	else
		@echo "Unknown OS: $(UNAME_S)"
		exit
	endif

	UNAME_M := $(shell uname -m)
	ifeq ($(UNAME_M),x86_64)
		HOSTARCH=amd64
	else ifneq ($(filter %86,$(UNAME_M)),)
		HOSTARCH=amd32
	else ifeq ($(UNAME_M),arm64)
		HOSTARCH=arm64
	else ifneq ($(filter arm%,$(UNAME_M)),)
		HOSTARCH=arm
	endif
endif

TARGETOS=$(HOSTOS)
ifdef MAKECMDGOALS
	ifeq ($(MAKECMDGOALS), $(filter $(MAKECMDGOALS), windows linux darwin))
		TARGETOS=$(MAKECMDGOALS)
	endif
endif

#GITTAG=$(shell git describe --tags $(shell git rev-list --tags --max-count=1))
GITCOMMIT=$(shell git log -1 --pretty=format:"%h")
BUILD_FLAGS=-a -ldflags "-w -s -X 'github.com/rigochain/rigo-go/cmd/version.GitCommit=$(GITCOMMIT)'"


BUILDDIR="./build/$(HOSTOS)"

all: pbm $(TARGETOS) tools

$(TARGETOS):
	@echo Build rigo for $(@) on $(HOSTOS)
ifeq ($(HOSTOS), windows)
	@set GOOS=$@& set GOARCH=$(HOSTARCH)& go build -o $(BUILDDIR)/rigo.exe $(BUILD_FLAGS)  ./cmd/
else
	@GOOS=$@ GOARCH=$(HOSTARCH) go build -o $(BUILDDIR)/rigo $(BUILD_FLAGS) ./cmd/
endif

pbm:
	@echo Compile protocol messages
	@protoc --go_out=$(GOPATH)/src -I./protos/ account.proto
	@protoc --go_out=$(GOPATH)/src -I./protos/ gov_params.proto
	@protoc --go_out=$(GOPATH)/src -I./protos/ trx.proto
	@protoc --go_out=$(GOPATH)/src -I./protos/ reward.proto

build-deploy:
	@echo "Build deploy tar file"

	@mkdir -p .deploy
	@mkdir -p .tmp/deploy
	@cp ./scripts/deploy/cli/files/* .tmp/deploy/
	@cp $(BUILDDIR)/rigo .tmp/deploy/

	@tar -czvf .deploy/deploy.gz.tar -C .tmp/deploy .
	@tar -tzvf .deploy/deploy.gz.tar
	@rm -rf .tmp/deploy

deploy:
	@echo Deploy...
	@sh -c scripts/deploy/deploy.sh

tools:
	@echo "Generate sfeeder.proto"
	@protoc --go_out=$(GOPATH)/src --go-grpc_out=$(GOPATH)/src -I./libs/sfeeder/protos secret_feeder.proto
	@echo "Build SecretFeeder ..."
	@go build -o $(BUILDDIR)/sfeeder -ldflags "-s -w" ./libs/sfeeder/server/cli/sfeeder.go
	@go build -o $(BUILDDIR)/sfsh -ldflags "-s -w" ./libs/sfeeder/client/cli/sfsh.go

check:
	@echo "OSTYPE: `echo ${OSTYPE}`"
	@echo "HOSTOS: $(HOSTOS)"
	@echo "HOSTARCH: $(HOSTARCH)"
	@echo "MAKECMDGOALS $(MAKECMDGOALS)"