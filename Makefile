.PHONY: test
test:
	GO111MODULE=on go test -v ./... -count 1

.PHONY: cover
cover:
	GO111MODULE=on go test -coverpkg=. -covermode=atomic -coverprofile=coverage.txt

.PHONY: view-cover
view-cover: cover
	GO111MODULE=on go tool cover -html coverage.txt

mod:
	GO111MODULE=on go mod tidy
