# First step, building binary
FROM golang:1.13.5-buster AS builder
WORKDIR /src
ENV CGO_ENABLED=0
ADD . .
RUN make build

# Second step, final container
FROM debian:buster-slim
COPY --from=builder /src/bin/go-kube-api /go-kube-api
ENTRYPOINT ["/go-kube-api"]