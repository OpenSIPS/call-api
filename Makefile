
BUILD_DIR ?= bin
CFG_BASE_DIR ?= /etc/call-api
BUILD_FLAGS ?= -i

CMD_TOOLS=$(wildcard cmd/*/main.go)

TOOLS?=$(patsubst cmd/%/main.go,%,$(CMD_TOOLS))

BUILD_TOOLS=$(addprefix $(BUILD_DIR)/,$(TOOLS))
INSTALL_TOOLS=$(addprefix $(GOPATH)/bin/%,$(TOOLS))

CFG_FILES=$(wildcard config/*.yml)
CFG_TOOLS=$(filter $(patsubst config/%.yml,%,$(CFG_FILES)),$(TOOLS))
INSTALL_CFG_TOOLS=$(addsuffix .yml,$(addprefix $(CFG_BASE_DIR)/,$(CFG_TOOLS)))

build: build-all

install: install-all

build-all: build-tools

install-all: install-tools install-cfg

install-cfg: $(INSTALL_CFG_TOOLS)

build-tools: $(BUILD_DIR) $(BUILD_TOOLS)

install-tools: $(GOPATH)/bin

$(BUILD_DIR) $(GOPATH)/bin $(CFG_BASE_DIR):
	@mkdir -p $@

$(BUILD_DIR)/%: cmd/%/main.go
	go build $(BUILD_FLAGS) -o $@ $^

$(GOPATH)/bin/%: cmd/%/main.go
	go install $(BUILD_FLAGS) -o $@ $^

$(CFG_BASE_DIR)/%.yml: config/%.yml $(CFG_BASE_DIR)
	@test -e $@ || cp $< $@

.PHONY: clean
clean:
	@rm -f $(BUILD_TOOLS)
