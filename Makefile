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

# Embed a build-time version if the user set one (fallback to 0.1.0 in code).
VERSION ?= $(shell git describe --tags --dirty 2>/dev/null)
LDFLAGS := -s -w $(if $(VERSION),-X main.version=$(VERSION),)

.PHONY: build install uninstall run clean tidy vet

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

tidy:
	go mod tidy

vet:
	go vet ./...
