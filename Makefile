.PHONY: check pins install standards sync-standards
SHELL := bash
ORG = https://github.com/lowmagic/.github
ENGINE = github.com/lowmagic/compa-engine
EXPERTS = github.com/lowmagic/compa-experts

# BINDIR is where make install puts compa; ~/.local/bin by default, which is
# usually on PATH.
BINDIR ?= $(HOME)/.local/bin

check: pins
	golangci-lint run
	go test -race ./...
	go mod tidy -diff
	git ls-files -z '*.sh' | xargs -0 -r shellcheck

# install builds compa from this checkout into BINDIR and says which build
# it is, so the compa on PATH is the one just built.
install:
	mkdir -p $(BINDIR)
	go build -o $(BINDIR)/compa .
	$(BINDIR)/compa version

# pins fails when the experts module this build requires names another
# compa-engine than this build does (go mod graph): Go builds both with the
# higher, so the experts would ship on an engine their own checks never ran.
pins:
	@export GOWORK=off && graph=$$(go mod graph) && main=$$(go list -m) && \
	pin() { awk -v from="$$1" -v to="$$2@" '$$1 == from && index($$2, to) == 1 { print substr($$2, length(to) + 1) }' <<<"$$graph"; } && \
	ours=$$(pin "$$main" $(ENGINE)) && experts=$$(pin "$$main" $(EXPERTS)) && \
	theirs=$$(pin "$(EXPERTS)@$$experts" $(ENGINE)) && \
	if [ -z "$$ours" ] || [ -z "$$experts" ] || [ "$$ours" != "$$theirs" ]; then \
		echo "pins: $$main requires $(ENGINE) $${ours:-(none)}, but $(EXPERTS) $${experts:-(none)} requires $${theirs:-(none)}"; \
		exit 1; \
	fi

standards:
	@tmp=$$(mktemp -d) && trap 'rm -rf "$$tmp"' EXIT && \
	git clone -q --depth 1 --branch main $(ORG) "$$tmp" && \
	"$$tmp/code-standards/sync.sh" check "$$tmp"

sync-standards:
	@tmp=$$(mktemp -d) && trap 'rm -rf "$$tmp"' EXIT && \
	git clone -q --depth 1 --branch main $(ORG) "$$tmp" && \
	"$$tmp/code-standards/sync.sh" sync "$$tmp" "$$(git -C "$$tmp" rev-parse HEAD)"
