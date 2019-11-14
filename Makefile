XGOARCH := arm
XGOOS := linux
XBINS := $(XGOOS)_$(XGOARCH)/wakeup $(XGOOS)_$(XGOARCH)/wakeupbr

.PHONY: $(XBINS)

all: deps test vet install

fmt:
	go fmt ./...

test:
	go test ./...

vet:
	go vet ./...

deps:
	go get -d -v ./...

install:
	go install ./...

xinstall:
	env GOOS=$(XGOOS) GOARCH=$(XGOARCH) go install ./...

publish: $(XBINS)

$(XBINS):
ifndef DEST_PATH
	$(error DEST_PATH must be set when publishing)
endif
	rsync -a $(GOPATH)/bin/$@ $(DEST_PATH)/$@
	@sha256sum $(GOPATH)/bin/$@
