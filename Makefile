CMD_DIR := ./cmd

BINS := $(notdir $(wildcard $(CMD_DIR)/*))

bin:
	@mkdir -p bin

build: bin
	@for bin in $(BINS); do \
    	echo "Building $$bin..."; \
		go build -o bin/$$bin $(CMD_DIR)/$$bin; \
	done

clean:
	@echo "Cleaning bins..."
	@rm bin/*

.PHONY: build clean bin
