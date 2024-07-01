# --- Go builder image ---
FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS build

ARG TARGETOS 
ARG TARGETARCH
ARG NUVLAEDGE_VERSION=dev
ARG GO_BUILD_TAGS

WORKDIR /build

COPY . .

#RUN apk add --no-cache tzdata
# RUN apk add --no-cache upx

RUN go mod tidy
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 \
    go build \
    -ldflags "-w -s -X 'nuvlaedge-go/nuvlaedge/version.Version=$NUVLAEDGE_VERSION'" \
    -gcflags=all="-l -B" \
    -o out/nuvlaedge \
    -tags "$GO_BUILD_TAGS" \
    ./cmd/nuvlaedge/



# --- Final image ---
FROM scratch

ENV PATH=/bin

# Add default certificates to allow HTTPS connections
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Template configuration for standard NuvlaEdge configuration
COPY --from=build /build/config/template.toml /etc/nuvlaedge/nuvlaedge.toml
# NuvlaEdge binary previously built
COPY --from=build /build/out/nuvlaedge /bin/nuvlaedge

ENTRYPOINT ["nuvlaedge"]
