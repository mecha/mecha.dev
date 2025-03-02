BUILD_TAGS="sqlite_fts5"

build:
	go build -tags $(BUILD_TAGS) .

dev:
	go run -tags $(BUILD_TAGS) . -w -v -p 8080

run:
	go run -tags $(BUILD_TAGS) . -v -t

test:
	go test -tags $(BUILD_TAGS) ./... -v
