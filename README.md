# go2rtc-tuya-oem

go2rtc runs `echo:` commands on every dial and reconnect, then dials the URL they print. This helper turns a saved white-label Tuya session into the `tuya://` URL, so the session id never sits in `go2rtc.yaml`. It is a static Go binary, because the official go2rtc image has no Go or Node.

## Session JSON

Uses the havent bridge config field names, plus a `ch_key`. havent does not write `ch_key` today, so it can come from the `--ch-key` flag instead. A `ch_key` in the JSON wins over the flag, and with neither the helper fails naming the field.

```json
{
  "signing_key": "...",
  "sid": "...",
  "ecode": "...",
  "partner": "...",
  "app_key": "...",
  "device_id": "phone device id",
  "package_name": "com.example.app",
  "api_host": "a1.tuyaeu.com",
  "ch_key": "...",
  "cameras": [{"camera_id": "abc123"}]
}
```

`api_host` may carry a scheme or path, only the host is used. It defaults to `a1.tuyaeu.com`. `talkback` is an optional boolean, false when missing. When true the URL ends with `&backchannel=1`, otherwise there is no `backchannel` param. A non-boolean value is an error. Other fields are ignored.

## Usage

```
go2rtc-tuya-oem --config /config/tuya-session.json [--ch-key <app ch key>] --camera abc123
```

Prints one URL on stdout:

```
tuya://<api_host>?device_id=<camera_id>&sid=..&ecode=..&partner_identity=..&app_key=..&signing_key=..&ch_key=..&package_name=..&client_device_id=..[&backchannel=1]
```

A missing field or unknown camera exits non-zero, names the field or camera on stderr and prints nothing on stdout. Secret values are never logged.

## go2rtc

```yaml
echo:
  allow_paths: [/usr/local/bin/go2rtc-tuya-oem]
streams:
  nursery: echo:/usr/local/bin/go2rtc-tuya-oem --config /config/tuya-session.json --ch-key <app ch key> --camera <id>
```

go2rtc refuses `echo:` sources added through its API, so this only works in the YAML file.

## Release

Pushing a tag `v*` builds static linux/amd64 and linux/arm64 binaries and attaches them to a GitHub release.

## Docker image

`ghcr.io/supergillis/go2rtc-tuya-oem:<tag>` and `:latest`, for linux/amd64 and linux/arm64, are published on every `v*` tag. The image is `alexxit/go2rtc` with go2rtc rebuilt from `supergillis/go2rtc` at the `GO2RTC_REF` build arg, plus `/usr/local/bin/go2rtc-tuya-oem`. Entrypoint, `/config` volume and default `/config/go2rtc.yaml` are unchanged, so the go2rtc docs apply.

```
docker run -d --name go2rtc --network host \
  -v /path/to/config:/config \
  ghcr.io/supergillis/go2rtc-tuya-oem:latest
```

```yaml
services:
  go2rtc:
    image: ghcr.io/supergillis/go2rtc-tuya-oem:latest
    network_mode: host
    volumes:
      - ./config:/config
```

Put `tuya-session.json` and a `go2rtc.yaml` with the `echo:` stream from the section above in `./config`.

The image reports the upstream go2rtc version it is based on, for example `1.9.15`, so Home Assistant sees no version mismatch. Only the revision in `go2rtc -version` shows the fork.

### Local build

Use a local go2rtc checkout instead of cloning. The directory root must be the go2rtc repo root and keep its `.git`.

```
podman build --build-context go2rtc-source=/path/to/go2rtc -t localhost/go2rtc-tuya-oem:dev .
podman run --rm localhost/go2rtc-tuya-oem:dev go2rtc -version
```

Without the override, the build clones `GO2RTC_REPO` at `GO2RTC_REF`. `GO2RTC_REF` defaults to a full commit sha in the Dockerfile. Bumping it is a deliberate commit, never a branch name.
