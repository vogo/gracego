
lint:
		golangci-lint run

fmt:
		go fmt

check: fmt lint