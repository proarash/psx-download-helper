## psx-helper

Local proxy that serves PS4/PS5 `.pkg` files from your PC so the console can install games you own over the LAN.

### Downloads

| OS      | x64                            | x86                          |
| ------- | ------------------------------ | ---------------------------- |
| Windows | `psx-helper-windows-amd64.exe` | `psx-helper-windows-386.exe` |
| Linux   | `psx-helper-linux-amd64`       | `psx-helper-linux-386`       |
| macOS   | `psx-helper-darwin-amd64`      | —                            |

Verify downloads with `SHA256SUMS.txt`.

### Quick start

```text
psx-helper --auto 1 --path <dir-with-pkg-files>
```

Set the console's proxy to the IP printed by the app, port `8080`. See README.md for all modes.
