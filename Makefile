GITV != git describe --tags
GITC != git rev-parse --verify HEAD
SRC  != find . -type f -name '*.go' ! -name '*_test.go'
TEST != find . -type f -name '*_test.go'

PREFIX  ?= /usr/local
VERSION ?= $(GITV)
COMMIT  ?= $(GITC)
BUILDER ?= Makefile
PKG     ?= github.com/makeworld-the-better-one/amfora/
SHARE   ?= /usr/share
LICENSE ?= $(SHARE)/licenses/amfora/LICENSE
THANKS  ?= $(SHARE)/doc/amfora/THANKS.md

GO      := go
INSTALL := install
RM      := rm

amfora: go.mod go.sum $(SRC)
	GO111MODULE=on CGO_ENABLED=0 $(GO) build -o $@ -ldflags="-s -w\
		-X main.version=$(VERSION) \
		-X main.commit=$(COMMIT) \
		-X main.builtBy=$(BUILDER) \
		-X $(PKG)display.licensePath=$(LICENSE) \
		-X $(PKG)display.thanksPath=$(THANKS) \
	"

.PHONY: clean
clean:
	$(RM) -f amfora

.PHONY: install
install: amfora amfora.desktop
	install -m 755 amfora $(PREFIX)/bin/amfora
	install -m 644 amfora.desktop $(PREFIX)/share/applications/amfora.desktop
	$(INSTALL) -Dm 644 LICENSE $(LICENSE)
	$(INSTALL) -Dm 644 THANKS.md $(THANKS)

.PHONY: uninstall
uninstall:
	$(RM) -f $(PREFIX)/bin/amfora
	$(RM) -f $(PREFIX)/share/applications/amfora.desktop
	$(RM) -rf $(SHARE)/licenses/amfora
	$(RM) -rf $(SHARE)/doc/amfora

# Development helpers
.PHONY: fmt
fmt:
	go fmt ./...
