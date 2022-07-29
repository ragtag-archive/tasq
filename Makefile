.PHONY: all
all: clean test tasq
	-upx tasq

tasq: $(shell find . -type f -name '*.go')
	go build -tags netgo -ldflags='-s -w -extldflags=-static' -o tasq

.PHONY: test
test:
	go test ./...

.PHONY: run
run: tasq
	./tasq

.PHONY: clean
clean:
	$(RM) tasq
