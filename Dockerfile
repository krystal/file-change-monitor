# Build the manager binary
FROM golang:1.13 as builder

WORKDIR /workspace
COPY go.sum go.sum
COPY go.mod go.mod
COPY main.go main.go
COPY pkg pkg
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o file-change-monitor main.go

FROM alpine:latest
WORKDIR /
COPY --from=builder /workspace/file-change-monitor /
CMD [ "/file-change-monitor" ]
