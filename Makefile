.PHONY: install
install:
	go get -u github.com/golang/dep/cmd/dep

.PHONY: init
init:
	dep ensure

.PHONY: test
test:
	go test -v ./...
