DATE=$(shell date -u +%Y-%m-%d)
VERSION=$(shell cat VERSION | sed 's/-dev//g')
DCGM_EXPORTER_VERSION=$(shell cat VERSION_DCGM_EXPORTER | sed 's/-dev//g')
DIST=$(shell lsb_release -c | cut -f2)

#########################################
# Tools                                 #
#########################################

TOOLS_DIR := hack/tools
include hack/tools.mk

#########################################
# Targets                                 #
#########################################

.PHONY: format
format: $(GOLICENSES) $(GOIMPORTS)
	@./hack/format.sh ./cmd ./pkg

.PHONY: test
test:
	@./hack/test.sh ./pkg/...

.PHONY: check
check: $(GOIMPORTS) $(GOLANGCI_LINT)
	@./hack/test.sh ./pkg/...
	@./hack/check.sh ./cmd/... ./pkg/...

.PHONY: build
build:
	@env GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -ldflags "-w -X github.com/digitalocean/do-dcgm-exporter/cmd/agent.version=${VERSION} -X github.com/digitalocean/do-dcgm-exporter/cmd/agent.dcgmExporterVersion=${DCGM_EXPORTER_VERSION} -X github.com/digitalocean/do-dcgm-exporter/cmd/agent.buildDate=${DATE}" -o bin/do-dcgm-exporter-linux-amd64 ./cmd/main.go

install: INSTALL ?= install
install: DESTDIR ?= debian/do-dcgm-exporter/
install:
	$(INSTALL) -s -D bin/do-dcgm-exporter-linux-amd64 $(DESTDIR)/opt/digitalocean/bin/do-dcgm-exporter
	mkdir -p $(DESTDIR)/etc/apt/trusted.gpg.d/
	mkdir -p $(DESTDIR)/etc/apt/sources.list.d/
	mkdir -p $(DESTDIR)/etc/systemd/system/
	cp public.gpg $(DESTDIR)/etc/apt/trusted.gpg.d/do-dcgm-exporter.gpg
	echo "deb https://digitalocean.github.io/do-dcgm-exporter/ubuntu/ $(DIST) extras" > $(DESTDIR)/etc/apt/sources.list.d/do-dcgm-exporter.list
	cp hack/systemd/do-dcgm-exporter.service $(DESTDIR)/etc/systemd/system/do-dcgm-exporter.service

debian/changelog:
	debian/doch.pl > debian/changelog

.PHONY: all
all: format check build

.PHONY: revendor
revendor:
	@GO111MODULE=on go mod vendor
	@GO111MODULE=on go mod tidy
