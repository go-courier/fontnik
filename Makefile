test:
	go test -v -race ./...

cover:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
