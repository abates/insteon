
test:
	go test ./...

coverage:
	go test -cover ./... -coverprofile=coverage.out

htmlcoverage: coverage
	go tool cover -html=coverage.out

funccoverage: coverage
	go tool cover -func=coverage.out

ic:
	go build ./cmd/ic

build: ic
