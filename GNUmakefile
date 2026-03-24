DISTDIR=${CURDIR}/dist
SPABALL=pkg/http/webroot.tar
SPADIR=${CURDIR}/frontend
SPADISTDIR=$(SPADIR)/dist/ed-survey-tools/browser

MOCKIDP_VERSION?=3.0.1
MOCKIDP_IMAGE=ghcr.io/navikt/mock-oauth2-server:$(MOCKIDP_VERSION)
MOCKIDP_PORT?=8080

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

$(SPABALL): $(shell find $(SPADIR)/src -type f)
	$(MAKE) -C $(SPADIR) build
	tar -C $(SPADISTDIR)/ -cvf $@ .

.PHONY: mock-idp
mock-idp:
	docker run -p $(MOCKIDP_PORT):8080 -it $(MOCKIDP_IMAGE)
