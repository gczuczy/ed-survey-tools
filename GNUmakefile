DISTDIR=${CURDIR}/dist
SPABALL=pkg/http/webroot.tar
SPADIR=${CURDIR}/frontend
SPADISTDIR=$(SPADIR)/dist/ed-survey-tools/browser

GOENV?=
BINSUFFIX=

NATIVEOS=$(shell go env GOOS)
NATIVEARCH=$(shell go env GOARCH)
GOOS?=$(NATIVEOS)
GOARCH?=$(NATIVEARCH)
ifdef GOOS
GOENV+= GOOS=$(GOOS)
endif
ifdef GOARCH
GOENV+= GOARCH=$(GOARCH)
endif
ifneq ($(GOOS)-$(GOARCH),$(NATIVEOS)-$(NATIVEARCH))
BINSUFFIX+=-$(GOOS)-$(GOARCH)
endif


SERVICE=$(DISTDIR)/edst$(BINSUFFIX)

sinclude $(HOME)/Mk/ed-survey-tools.mk

$(DISTDIR):
	mkdir -p $@

.PHONY: cli
cli: $(DISTDIR)/sdsheet

go.mod: $(shell find ./ -type f -name '*.go')
	go mod tidy

.PHONY: build
build: $(SERVICE)
	@echo "built stuff"

$(SERVICE): go.mod $(SPABALL) $(shell find ./ -type f -name '*.go') | $(DISTDIR)
	$(GOENV) CGO_ENABLED=0 go build -o $@  .

.PHONY: frontend
frontend: $(SPABALL)

$(SPABALL): $(shell find $(SPADIR)/src -type f) $(SPADIR)/package-lock.json
	$(MAKE) -C $(SPADIR) build
	tar -C $(SPADISTDIR)/ -cvf $@ .
