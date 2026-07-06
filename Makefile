test:
	go test ./...

lint:
	golangci-lint run

lint-fix:
	golangci-lint run --fix
