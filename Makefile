# codearts-cli — build / install / uninstall
#
# Usage:
#   make build                            # build ./codearts-cli in repo
#   make install                          # install to /usr/local/bin  (may need sudo)
#   make install PREFIX=$HOME/.local      # install to ~/.local/bin    (no sudo)
#   make uninstall                        # remove installed binary
#   make run ARGS="pipeline run <id>"     # build + run

BINARY ?= codearts-cli
PREFIX ?= /usr/local
BINDIR := $(PREFIX)/bin
PKG    := .

# Embed a build-time version (fallback is "dev" in cmd/root.go).
VERSION ?= $(shell git describe --tags --dirty 2>/dev/null)
LDFLAGS := -s -w $(if $(VERSION),-X github.com/Lzhtommy/codearts-cli/cmd.version=$(VERSION),)

.PHONY: build install uninstall run clean tidy vet dist npm-pack npm-link

build:
	go build -ldflags '$(LDFLAGS)' -o $(BINARY) $(PKG)
	@echo "✓ built ./$(BINARY)"

install: build
	@mkdir -p $(BINDIR)
	@install -m 0755 $(BINARY) $(BINDIR)/$(BINARY)
	@echo "✓ installed $(BINDIR)/$(BINARY)"
	@case ":$$PATH:" in \
	  *":$(BINDIR):"*) echo "✓ $(BINDIR) is on PATH — run: $(BINARY) --help" ;; \
	  *) echo "! $(BINDIR) is NOT on PATH. Add it, e.g.:"; \
	     echo "    echo 'export PATH=\"$(BINDIR):\$$PATH\"' >> ~/.zshrc && source ~/.zshrc" ;; \
	esac

uninstall:
	@rm -f $(BINDIR)/$(BINARY)
	@echo "✓ removed $(BINDIR)/$(BINARY)"

run: build
	./$(BINARY) $(ARGS)

clean:
	rm -f $(BINARY)
	rm -rf dist/

tidy:
	go mod tidy

vet:
	go vet ./...

# ---------- npm / cross-compile ----------

DIST := dist
PLATFORMS := darwin-amd64 darwin-arm64 linux-amd64 linux-arm64 windows-amd64 windows-arm64
NPM_VERSION := $(shell node -p "require('./package.json').version")

# Build archives for all platforms (used by GitHub Releases / npm postinstall).
# Produces dist/codearts-cli-{version}-{os}-{arch}.tar.gz (or .zip for windows).
dist:
	@mkdir -p $(DIST)
	@for target in $(PLATFORMS); do \
		goos=$$(echo $$target | cut -d- -f1); \
		goarch=$$(echo $$target | cut -d- -f2); \
		suffix=""; [ "$$goos" = "windows" ] && suffix=".exe"; \
		outbin="$(DIST)/$(BINARY)$$suffix"; \
		echo "building $$goos/$$goarch ..."; \
		CGO_ENABLED=0 GOOS=$$goos GOARCH=$$goarch \
			go build -ldflags '$(LDFLAGS)' -o "$$outbin" $(PKG); \
		if [ "$$goos" = "windows" ]; then \
			(cd $(DIST) && zip -q "$(BINARY)-$(NPM_VERSION)-$$goos-$$goarch.zip" "$(BINARY)$$suffix"); \
		else \
			tar -czf "$(DIST)/$(BINARY)-$(NPM_VERSION)-$$goos-$$goarch.tar.gz" -C $(DIST) "$(BINARY)$$suffix"; \
		fi; \
		rm -f "$$outbin"; \
	done
	@echo "✓ archives in $(DIST)/"

# Pack an npm tarball (for local testing / private registry).
npm-pack:
	npm pack
	@echo "✓ npm tarball created"

# npm link — installs the package globally from the working tree.
# Useful for local dev: builds the native binary into bin/ then links.
npm-link: build
	@mkdir -p bin
	@cp $(BINARY) bin/$(BINARY)
	npm link
	@echo "✓ npm link done — codearts-cli is now globally available"
