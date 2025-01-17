ARG IMAGE_PREFIX
ARG GO_IMAGE=golang:1.23.0

# Build the manager binary
FROM ${IMAGE_PREFIX}${GO_IMAGE} AS builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY Makefile Makefile
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN make mod-download

# Copy the go source
COPY . .

# Build ework
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make build

# Use distroless as minimal base image
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot

WORKDIR /

COPY --from=builder /workspace/bin/tgbot .
COPY --from=builder /workspace/.env .
USER 65532:65532

ENTRYPOINT ["./tgbot", "poll"]
