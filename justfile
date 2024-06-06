default: build

build:
    go build ./...

test:
    go test ./...

lint:
    act -j lint

verify: lint
    act -j build

readme:
    act -j validate_links