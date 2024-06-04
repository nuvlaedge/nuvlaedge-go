# --- Go builder image ---
FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS build

ARG TARGETOS 
ARG TARGETARCH

WORKDIR /build

COPY . .

#RUN apk add --no-cache tzdata
RUN apk add --no-cache upx

RUN go mod tidy
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build -o out/nuvlaedge ./cmd/nuvlaedge.go
RUN upx --lzma out/nuvlaedge


# --- NuvlaEdge image ---
FROM sixsq/nuvlaedge AS nuvlaedge


# --- Final image ---
FROM scratch

ENV NUVLA_ENDPOINT=https://nuvla.io \
    DATA_LOCATION=/var/lib/nuvlaedge/ \
    HEARTBEAT_PERIOD=20 \
    TELEMETRY_PERIOD=60 \
    PATH=/bin

#COPY --from=build /etc/passwd /etc/passwd
#COPY --from=build /etc/group /etc/group
#COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=build /build/out/nuvlaedge /bin/nuvlaedge

COPY --from=nuvlaedge /usr/local/libexec/docker/cli-plugins/docker-compose /bin/docker-compose

ENTRYPOINT ["nuvlaedge"]
CMD []
