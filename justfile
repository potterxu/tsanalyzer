default: build

build:
    CGO_ENABLED=0 go build -ldflags '-w -extldflags "-static"'

test:
    go test ./...

lint:
    act -j lint

verify: lint
    act -j build

readme:
    act -j validate_links