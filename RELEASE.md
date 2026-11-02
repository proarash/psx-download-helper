## psx-helper

Local proxy that serves PS4/PS5 `.pkg` files from your PC so the console can install games you own over the LAN.

### Downloads

| OS | x64 | x86 |
| --- | --- | --- |
| Windows | [psx-helper-windows-amd64.exe](https://github.com/proarash/psx-download-helper/releases/latest/download/psx-helper-windows-amd64.exe) | [psx-helper-windows-386.exe](https://github.com/proarash/psx-download-helper/releases/latest/download/psx-helper-windows-386.exe) |
| Linux | [psx-helper-linux-amd64](https://github.com/proarash/psx-download-helper/releases/latest/download/psx-helper-linux-amd64) | [psx-helper-linux-386](https://github.com/proarash/psx-download-helper/releases/latest/download/psx-helper-linux-386) |
| macOS | [psx-helper-darwin-amd64](https://github.com/proarash/psx-download-helper/releases/latest/download/psx-helper-darwin-amd64) | — |

Verify downloads with [SHA256SUMS.txt](https://github.com/proarash/psx-download-helper/releases/latest/download/SHA256SUMS.txt).

### How to build

Requires Go. Binaries are written to `out/`.

```text
./build.ps1   # Windows
./build.sh    # Linux/macOS
```

### Program modes

| Command | What it does |
| --- | --- |
| `psx-helper --auto 0 --path <dir>` | Listens on your Ethernet IP, port `8080`, logs `.pkg` URLs, and serves local `.pkg` files |
| `psx-helper --auto 1 --path <dir>` | Same as `--auto 0`, and also saves every part URL of each game to `<package>_0.txt` |
| `psx-helper --addr :8080 --path <dir>` | Plain proxy on any address that serves local `.pkg` files |
| `psx-helper --get-urls <url_0.pkg>` | Finds all parts (`_0`, `_1`, …) and saves them to a `.txt` file |

### Console proxy tutorial

1. Run `psx-helper --auto 1 --path D:\pkgs` and note the IP it prints.
2. On the PS4/PS5, open **Settings → Network → Set Up Internet Connection**. Choose your connection, then **Custom** (PS4) or **Advanced Settings** (PS5).
3. Leave everything on automatic except **Proxy Server**: set it to **Use**, enter the printed IP as the address and `8080` as the port.
4. Start the game download on the console. The app prints the URL and creates `<package>_0.txt`.
5. Download every link in the `.txt` into `D:\pkgs` with a download manager. Keep the original file names.
6. Retry the download on the console. Files found in `D:\pkgs` are installed from the PC.
7. When you're done, set **Proxy Server** back to **Do Not Use**.
