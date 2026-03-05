.PHONY: fmt test check run-engine run-cli

fmt:
	go -C backend/engine fmt ./...
	go -C backend/cli fmt ./...

test:
	go -C backend/engine test ./...
	go -C backend/cli test ./...

check:
	go version
	go work sync
	$(MAKE) fmt
	$(MAKE) test

run-engine:
	go run ./backend/engine/cmd/engine

run-cli:
	go run ./backend/cli/cmd/nebula-cli --help
