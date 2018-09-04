.PHONY: test
test:
	GO111MODULE=on go test -v ./...

.PHONY: cover
cover:
	GO111MODULE=on go test -coverpkg=. -covermode=atomic -coverprofile=coverage.txt

.PHONY: view-cover
view-cover: cover
	GO111MODULE=on go tool cover -html coverage.txt

