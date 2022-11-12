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

BUILDDIR="./build/$(HOSTOS)"

all: $(TARGETOS)

$(TARGETOS):
	@echo Build arcanus for $(@) on $(HOSTOS)
ifeq ($(HOSTOS), windows)
	@set GOOS=$@& set GOARCH=$(HOSTARCH)& go build -o $(BUILDDIR)/arcanus.exe ./cmd/
else
	@GOOS=$@ GOARCH=$(HOSTARCH) go build -o $(BUILDDIR)/arcanus ./cmd/
endif

pbm:
	@echo Compile protocol messages
	@protoc --go_out=$(GOPATH)/src -I./protos/ account.proto
	@protoc --go_out=$(GOPATH)/src -I./protos/ govrule.proto
	@protoc --go_out=$(GOPATH)/src -I./protos/ trx.proto

build-deploy:
	@echo "Build deploy tar file"

	@mkdir -p .deploy
	@mkdir -p .tmp/deploy
	@cp ./scripts/deploy/cli/files/* .tmp/deploy/
	@cp $(BUILDDIR)/arcanus .tmp/deploy/

	@tar -czvf .deploy/deploy.gz.tar -C .tmp/deploy .
	@tar -tzvf .deploy/deploy.gz.tar
	@rm -rf .tmp/deploy

deploy:
	@echo Deploy...
	@sh -c scripts/deploy/deploy.sh

check:
	@echo "OSTYPE: `echo ${OSTYPE}`"
	@echo "HOSTOS: $(HOSTOS)"
	@echo "HOSTARCH: $(HOSTARCH)"
	@echo "MAKECMDGOALS $(MAKECMDGOALS)"