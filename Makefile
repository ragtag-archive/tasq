tasq: $(shell find . -type f -name '*.go')
	go build -tags netgo -ldflags='-s -w -extldflags=-static' -o tasq
	-upx tasq

.PHONY: run
run: tasq
	./tasq

.PHONY: clean
clean:
	$(RM) tasq
