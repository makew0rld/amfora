GIT_EXISTS := $(shell git status > /dev/null 2>&1 ; echo $$?)

ifeq ($(GIT_EXISTS), 0)
	ifeq ($(shell git tag --points-at HEAD),)
		# Not currently on a tag
		VERSION := $(shell git describe --tags | sed 's/-.*/-next/') # v1.2.3-next
	else
		# On a tag
		VERSION := $(shell git tag --points-at HEAD)
	endif

	COMMIT := $(shell git rev-parse --verify HEAD)
endif

INSTALL := install -o root -g root
INSTALL_DIR := /usr/local/bin
DESKTOP_DIR := /usr/share/applications

.PHONY: all build install desktop clean uninstall fmt

all: build

build:
ifneq ($(GIT_EXISTS), 0)
	# No Git repo
	$(error No Git repo was found, which is needed to compile the commit and version)
endif
	@echo "Downloading dependencies"
	@go env -w GO111MODULE=on ; go mod download
	@echo "Building binary"
	@go env -w GO111MODULE=on CGO_ENABLED=0 ; go build -ldflags="-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.builtBy=Makefile"

install:
	@echo "Installing Amfora to $(INSTALL_DIR)"
	@$(INSTALL) -m 755 amfora $(INSTALL_DIR)

desktop:
	@echo "Setting up desktop file"
	@$(INSTALL) -m 644 amfora.desktop $(DESKTOP_DIR)
	@update-desktop-database $(DESKTOP_DIR)

clean:
	@echo "Removing Amfora binary in local directory"
	@$(RM) amfora

uninstall:
	@echo "Removing Amfora from $(INSTALL_DIR)"
	@$(RM) $(INSTALL_DIR)/amfora
	@echo "Removing desktop file"
	-@$(RM) $(DESKTOP_DIR)/amfora.desktop
	-@update-desktop-database $(DESKTOP_DIR)

fmt:
	go fmt ./...
