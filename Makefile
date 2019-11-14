XGOARCH := arm
XGOOS := linux
XBINS := $(XGOOS)_$(XGOARCH)/wakeup $(XGOOS)_$(XGOARCH)/wakeupbr

.PHONY: $(XBINS)

all: lint test install

test:
	go test ./...

vet:
	go vet ./...

check-fmt:
	bash -c "diff --line-format='%L' <(echo -n) <(gofmt -d -s .)"

lint: check-fmt vet

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
