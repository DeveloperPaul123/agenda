default:
    @just -l

[group('dev')]
[doc('Build the project')]
build:
    go build


[group('dev')]
[doc('Format the code')]
fmt:
    gofmt -w .

[group('dev')]
[doc('Lint the code')]
lint:
    golangci-lint run