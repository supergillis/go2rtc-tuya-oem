ARG GO_VERSION="1.27"
ARG GO2RTC_BASE_VERSION="1.9.15"

# The pinned commit is a deliberate choice. Bump GO2RTC_REF in its own commit.
FROM --platform=$BUILDPLATFORM docker.io/library/alpine:3 AS go2rtc-clone
ARG GO2RTC_REPO="https://github.com/supergillis/go2rtc.git"
ARG GO2RTC_REF="19cabe3d43ed6c1ea76dda9f1dc16d15899792ca"
RUN apk add --no-cache git
WORKDIR /src
RUN git init -q . \
    && git fetch -q --depth 1 "${GO2RTC_REPO}" "${GO2RTC_REF}" \
    && git checkout -q FETCH_HEAD

# Replaced by `--build-context go2rtc-source=<dir>` for local builds. The
# directory root must be the go2rtc repo root with its .git kept, so that
# go2rtc -version reports the revision.
FROM scratch AS go2rtc-source
COPY --from=go2rtc-clone /src /

FROM --platform=$BUILDPLATFORM docker.io/library/golang:${GO_VERSION}-alpine AS go2rtc-build
ARG TARGETOS
ARG TARGETARCH
ARG GO2RTC_BASE_VERSION
ENV GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0
# git is required for the VCS stamp that go2rtc -version prints.
RUN apk add --no-cache git
WORKDIR /build
COPY --from=go2rtc-source / ./
# Tagging HEAD v<app.Version> makes Go stamp a matching module version, so
# go2rtc does not append +dev and Home Assistant sees the clean version.
# The tag is forced because a full checkout already carries upstream tags.
# The version is read from the source, never hardcoded.
RUN version="$(sed -n 's/.*app\.Version = "\(.*\)".*/\1/p' main.go)" \
    && test -n "${version}" \
    && { test "${version}" = "${GO2RTC_BASE_VERSION}" || { echo "go2rtc source is ${version} but the base image is ${GO2RTC_BASE_VERSION}, bump GO2RTC_BASE_VERSION too" >&2; exit 1; }; } \
    && git -c safe.directory='*' tag -f "v${version}" \
    && go build -ldflags "-s -w" -trimpath

FROM --platform=$BUILDPLATFORM docker.io/library/golang:${GO_VERSION}-alpine AS helper-build
ARG TARGETOS
ARG TARGETARCH
ENV GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0
WORKDIR /build
COPY go.mod ./
COPY *.go ./
RUN go build -ldflags "-s -w" -trimpath -o go2rtc-tuya-oem .

# Same layout as alexxit/go2rtc: tini entrypoint, /config volume, same CMD.
FROM docker.io/alexxit/go2rtc:${GO2RTC_BASE_VERSION}
COPY --from=go2rtc-build /build/go2rtc /usr/local/bin/go2rtc
COPY --from=helper-build /build/go2rtc-tuya-oem /usr/local/bin/go2rtc-tuya-oem
