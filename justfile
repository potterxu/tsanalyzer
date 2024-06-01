default: build

build:
    go build ./...

test:
    go test ./...

lint:
    act -j lint