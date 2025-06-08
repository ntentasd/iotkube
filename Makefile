CMD_DIR := ./cmd
BINS := $(notdir $(wildcard $(CMD_DIR)/*))

bin:
	@mkdir -p bin

completions:
		@mkdir -p completions

build: bin completions
	@for bin in $(BINS); do \
    	echo "Building $$bin..."; \
		go build -o bin/$$bin $(CMD_DIR)/$$bin; \
	done

buildall: build
	@for bin in $(BINS); do \
		echo "Generating Zsh completion for $$bin..."; \
		./bin/$$bin completion zsh > completions/_$$bin; \
	done
	@$(MAKE) reload-completions

clean:
	@echo "Cleaning bins and completions..."
	@rm -f bin/* completions/_*

reload-completions:
	@echo "Reloading Zsh completions..."
	@rm -f ~/.zcompdump
	@zsh -ic 'autoload -Uz compinit && compinit'

.PHONY: build buildall clean bin completions reload-completions
