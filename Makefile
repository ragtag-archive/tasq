tasq: $(shell find . -type f -name '*.go')
	go build

.PHONY: run
run: tasq
	./tasq
