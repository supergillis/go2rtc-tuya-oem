ARG GO_VERSION="1.27"
ARG GO2RTC_IMAGE="ghcr.io/supergillis/go2rtc:1.9.15-oem.1"

FROM --platform=$BUILDPLATFORM docker.io/library/golang:${GO_VERSION}-alpine AS helper-build
ARG TARGETOS
ARG TARGETARCH
ENV GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0
WORKDIR /build
COPY go.mod ./
COPY *.go ./
RUN go build -ldflags "-s -w" -trimpath -o go2rtc-tuya-oem .

# Entrypoint, volume and CMD come from the fork image unchanged.
FROM ${GO2RTC_IMAGE}
COPY --from=helper-build /build/go2rtc-tuya-oem /usr/local/bin/go2rtc-tuya-oem
