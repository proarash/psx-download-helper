$ErrorActionPreference = 'Stop'

Set-Location $PSScriptRoot

$name = 'psx-helper'
$targets = @('windows/amd64', 'linux/amd64', 'darwin/amd64')

if (Test-Path out) { Remove-Item -Recurse -Force out }
New-Item -ItemType Directory -Force out | Out-Null

$env:CGO_ENABLED = '0'
try {
    foreach ($t in $targets) {
        $os, $arch = $t -split '/'
        $ext = if ($os -eq 'windows') { '.exe' } else { '' }
        $output = "out/$name-$os-$arch$ext"
        Write-Host "building $output"
        $env:GOOS = $os
        $env:GOARCH = $arch
        go build -trimpath -ldflags="-s -w" -o $output .
        if ($LASTEXITCODE -ne 0) { throw "build failed for $t" }
    }
}
finally {
    Remove-Item Env:GOOS, Env:GOARCH, Env:CGO_ENABLED -ErrorAction SilentlyContinue
}
