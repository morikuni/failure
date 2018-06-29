.PHONY: install
install:
	go get -u github.com/golang/dep/cmd/dep

.PHONY: init
init:
	dep ensure

.PHONY: test
test:
	go test -v ./...

.PHONY: cover
cover:
	go test -coverpkg=. -covermode=atomic -coverprofile=coverage.txt

.PHONY: view-cover
view-cover: cover
	go tool cover -html coverage.txt

