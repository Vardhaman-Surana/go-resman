FROM golang:1.12 AS builder
WORKDIR /go/src/github.com/vds/go-resman
ENV GO111MODULE=on
COPY ./go.mod ./go.sum ./
RUN go mod download
RUN go mod verify
COPY ./ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o ./bin/server ./cmd/main.go

FROM alpine:latest AS restaurant-management-server
RUN apk --no-cache add ca-certificates
RUN apk add --no-cache bash
COPY --from=builder /go/src/github.com/vds/go-resman/database/. ./database/
COPY --from=builder /go/src/github.com/vds/go-resman/bin/server .
CMD ["./server"]
