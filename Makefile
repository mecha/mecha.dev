BUILD_TAGS="sqlite_fts5"
LDFLAGS="-X main.Version=$$(git rev-parse --short HEAD)"

.PHONY: build dev test

build:
	go build -tags $(BUILD_TAGS) -ldflags $(LDFLAGS) .

dev:
	go run -tags $(BUILD_TAGS) -ldflags $(LDFLAGS) . -verbose -noembed -watch 

test:
	go test -tags $(BUILD_TAGS) ./... -v
