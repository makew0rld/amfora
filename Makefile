GITV != git describe --tags
GITC != git rev-parse --verify HEAD
SRC  != find . -type f -name '*.go' ! -name '*_test.go'
TEST != find . -type f -name '*_test.go'

PREFIX  ?= /usr/local
VERSION ?= $(GITV)
COMMIT  ?= $(GITC)
BUILDER ?= Makefile

GO      := go
INSTALL := install
RM      := rm

amfora: go.mod go.sum $(SRC)
	GO111MODULE=on CGO_ENABLED=0 $(GO) build -o $@ -ldflags="-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.builtBy=$(BUILDER)"

.PHONY: clean
clean:
	$(RM) -f amfora

.PHONY: install
install: amfora amfora.desktop
	$(INSTALL) -d $(DESTDIR)$(PREFIX)/bin/
	$(INSTALL) -m 755 amfora $(DESTDIR)$(PREFIX)/bin/amfora
	$(INSTALL) -d $(DESTDIR)$(PREFIX)/share/applications/
	$(INSTALL) -m 644 amfora.desktop $(DESTDIR)$(PREFIX)/share/applications/amfora.desktop

.PHONY: uninstall
uninstall:
	$(RM) -f $(DESTDIR)$(PREFIX)/bin/amfora
	$(RM) -f $(DESTDIR)$(PREFIX)/share/applications/amfora.desktop

# Development helpers

.PHONY: fmt
fmt:
	$(GO) fmt ./...

.PHONY: gen
gen:
	$(GO) generate ./...
