# Set variables using shell commands
VERSION := `git describe --tags --abbrev=0 2>/dev/null || echo "dev"`
COMMIT := `git rev-parse --short HEAD 2>/dev/null || echo "none"`

default:
    @just -l

[group('dev')]
[doc('Build the project')]
build:
    go build -ldflags "-X main.version={{VERSION}} -X main.commit={{COMMIT}}"


[group('dev')]
[doc('Format the code')]
fmt:
    gofmt -w .

[group('dev')]
[doc('Lint the code')]
lint:
    golangci-lint run

[group('dev')]
[doc('Run tests')]
test:
    go test ./...
