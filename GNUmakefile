DISTDIR=${CURDIR}/dist
SPABALL=pkg/http/webroot.tar
SPADIR=${CURDIR}/frontend
SPADISTDIR=$(SPADIR)/dist/ed-survey-tools/browser

SERVICE=$(DISTDIR)/edst

MOCKIDP_VERSION?=3.0.1
MOCKIDP_IMAGE=ghcr.io/navikt/mock-oauth2-server:$(MOCKIDP_VERSION)
MOCKIDP_PORT?=8080

$(DISTDIR):
	mkdir -p $@

.PHONY: cli
cli: $(DISTDIR)/sdsheet

go.mod: $(shell find ./ -type f -name '*.go')
	go mod tidy

$(DISTDIR)/sdsheetscraper: $(DISTDIR) go.mod $(shell find ./ -type f -name '*.go')
	CGO_ENABLED=0 go build -C cmd/cli/sdsheetscraper -o $@  .

.PHONY: build
build: $(SERVICE)
	@echo "built stuff"

$(SERVICE): go.mod $(SPABALL) $(shell find ./ -type f -name '*.go') | $(DISTDIR)
	CGO_ENABLED=0 go build -o $@  .

.PHONY: frontend
frontend: $(SPABALL)

$(SPABALL): $(shell find $(SPADIR)/src -type f)
	$(MAKE) -C $(SPADIR) build
	tar -C $(SPADISTDIR)/ -cvf $@ .

.PHONY: mock-idp
mock-idp:
	docker run -p $(MOCKIDP_PORT):8080 -it $(MOCKIDP_IMAGE)
