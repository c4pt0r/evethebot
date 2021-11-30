default: build

build: export GO111MODULE=on
build:
ifeq ($(TAGS),)
	$(CGO_FLAGS) go build -gcflags=-G=3 -tags json1 -o bin/evebot ./*.go
else
	$(CGO_FLAGS) go build -gcflags=-G=3 -tags json1 -tags "$(TAGS)" -o bin/evebot ./*.go
endif

check:
	golint -set_exit_status .

