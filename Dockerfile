FROM golang:1.24

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -tags sqlite_fts5 -v -o /usr/local/bin/main .

EXPOSE 8080
CMD ["main", "-v", "-p", "8080"]
