# psx-download-helper

A local HTTP proxy for PS4/PS5 that serves `.pkg` game files from your PC's disk instead of the internet. It helps you install games you own a license for when your internet connection is slow: download the package files once on the PC (with any download manager), then let the console install them over your LAN.

Written in Go using only the standard library.

## Build

```powershell
./build.ps1   # Windows
./build.sh    # Linux/macOS
```

Builds stripped binaries into `out/` for windows/amd64, windows/386, linux/amd64, linux/386, and darwin/amd64.

## Release

Push a tag like `v1.0.0` (or run the `release` workflow manually with a tag). GitHub Actions builds all targets with `build.sh` and publishes them, plus `SHA256SUMS.txt`, to a GitHub release. [RELEASE.md](RELEASE.md) is used as the release notes.

## Modes

### Auto mode (recommended)

```text
psx-helper --auto 1 --path D:\pkgs
```

- Listens on your **Ethernet** adapter's IPv4 address, port `8080`. Virtual, VPN, and WSL adapters are skipped.
- Serves any requested `.pkg` file found in `--path` (default: current directory), with range/resume support.
- `--auto 0`: only logs the `.pkg` URLs the console requests.
- `--auto 1`: logs them too, and for each new package it automatically probes every part (`_0.pkg`, `_1.pkg`, …) and writes the list to `<package>.txt` inside `--path`.

### Manual proxy mode

```text
psx-helper --addr :8080 --path D:\pkgs
```

Same serving behaviour as auto mode, on any address you choose. It doesn't log or generate link lists.

### Link list generation

```text
psx-helper --get-urls "http://.../EP9000-CUSA00001_00-XXXX_0.pkg" [--out game.txt]
```

Checks parts `_0`, `_1`, `_2`, … concurrently until one stops responding, then writes the working URLs, one per line, to a `.txt` file.

## Workflow

1. Run `--auto 1`.
2. On the console, go to network settings and set a manual proxy: server = the IP printed by the app, port = `8080`.
3. Start the game download on the console. The app logs the URL and creates the `.txt` link list.
4. Download every link in the `.txt` with a download manager into `--path`. Keep the original file names.
5. Restart the download on the console. Files found locally are served from disk; anything else passes through to the internet.

## Notes

- Interception only works for plain-HTTP `.pkg` requests. HTTPS traffic is tunnelled untouched.
- Local files are matched by file name only.
