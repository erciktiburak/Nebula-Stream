.PHONY: check

check:
	go version
	go work sync
	go test ./backend/engine/...
	go test ./backend/cli/...
