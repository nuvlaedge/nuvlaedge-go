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
    -ldflags "-w -s -X 'nuvlaedge-go/common/version.Version=$NUVLAEDGE_VERSION'" \
    -gcflags=all="-l -B" \
    -o out/nuvlaedge \
    -tags "$GO_BUILD_TAGS" \
    ./cmd/nuvlaedge/



# --- Final image ---
FROM scratch

ARG GIT_BRANCH
ARG GIT_COMMIT_ID
ARG GIT_BUILD_TIME
ARG GITHUB_RUN_NUMBER
ARG GITHUB_RUN_ID
ARG PROJECT_URL

LABEL git.branch=${GIT_BRANCH} \
      git.commit.id=${GIT_COMMIT_ID} \
      git.build.time=${GIT_BUILD_TIME} \
      git.run.number=${GITHUB_RUN_NUMBER} \
      git.run.id=${GITHUB_RUN_ID}
LABEL org.opencontainers.image.authors="support@sixsq.com" \
      org.opencontainers.image.created=${GIT_BUILD_TIME} \
      org.opencontainers.image.url=${PROJECT_URL} \
      org.opencontainers.image.vendor="SixSq SA" \
      org.opencontainers.image.title="NuvlaEdge" \
      org.opencontainers.image.description="NuvlaEdge agent in Golang"

ENV PATH=/bin

# Add default certificates to allow HTTPS connections
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# NuvlaEdge binary previously built
COPY --from=build /build/out/nuvlaedge /bin/nuvlaedge

ENTRYPOINT ["nuvlaedge"]
