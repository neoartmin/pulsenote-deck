#requires -Version 5.1
<#
Import the previously built Windows v7.5.8 all-in-one ZIP to the SAME
PulseNote repository while preserving every path inside the ZIP.

Usage (from a clone on the feature/pn15-grip branch):
  powershell -ExecutionPolicy Bypass -File .\tools\import-windows-v758.ps1 -ArchivePath "C:\Downloads\PND_Windows_v7.5.8_PN12_PN15_3Devices_전체파일집.zip"

This imports the binary installer and complete Go sources (including icons),
plus the three firmware copies found in the Windows package.
#>
param(
    [Parameter(Mandatory=$true)]
    [string]$ArchivePath,
    [switch]$SkipPush
)
$ErrorActionPreference = "Stop"
if (!(Test-Path -LiteralPath $ArchivePath -PathType Leaf)) {
    throw "Archive not found: $ArchivePath"
}
$expected = "PND_Windows_v7.5.8_PN12_PN15_3Devices_전체파일집"
$archiveSha256 = "270f8839b092ca8726f159fd9474a440cb305d56cb687434cc1a2fb59b7dca25"
$actualSha256 = (Get-FileHash -LiteralPath $ArchivePath -Algorithm SHA256).Hash.ToLowerInvariant()
if ($actualSha256 -ne $archiveSha256) {
    throw "SHA256 mismatch: Windows archive is not the verified v7.5.8 package"
}
$gitRoot = (& git rev-parse --show-toplevel).Trim()
if ($LASTEXITCODE -ne 0 -or !$gitRoot) { throw "Run from a Git clone" }
$remote = (& git -C $gitRoot remote get-url origin).Trim()
if ($remote -notmatch 'neoartmin/pulsenote-deck(\.git)?$') {
    throw "Wrong repository remote: $remote"
}
$branch = (& git -C $gitRoot branch --show-current).Trim()
if ($branch -ne "feature/pn15-grip") {
    throw "Switch to feature/pn15-grip before importing. Current: $branch"
}
$changes = (& git -C $gitRoot status --porcelain)
if ($changes) { throw "Working tree is not clean. Commit or stash changes first." }

$tmp = Join-Path ([System.IO.Path]::GetTempPath()) ("PND-import-" + [guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Path $tmp | Out-Null
try {
    Expand-Archive -LiteralPath $ArchivePath -DestinationPath $tmp -Force
    $sourceRoot = Join-Path $tmp $expected
    $required = @(
        "소스/app/go.mod",
        "소스/app/gui_windows.go",
        "소스/installer/go.mod",
        "소스/installer/main_windows.go",
        "설치/PulseNoteDeck_v7.5.8_Setup.exe",
        "펌웨어/PN12/PulseNoteDeck_PN12.ino",
        "펌웨어/PN15_LEFT/PulseNoteDeck_PN15_LEFT.ino",
        "펌웨어/PN15_RIGHT/PulseNoteDeck_PN15_RIGHT.ino",
        "SHA256SUMS.txt"
    )
    foreach ($rel in $required) {
        if (!(Test-Path -LiteralPath (Join-Path $sourceRoot $rel) -PathType Leaf)) {
            throw "Invalid package: $rel is absent"
        }
    }
    $target = Join-Path $gitRoot "Windows"
    if (Test-Path -LiteralPath $target) {
        Remove-Item -LiteralPath $target -Force -Recurse
    }
    New-Item -ItemType Directory -Path $target | Out-Null
    # Preserve the original subfolder names and every file's relative path.
    Get-ChildItem -LiteralPath $sourceRoot -Force | ForEach-Object {
        Copy-Item -LiteralPath $_.FullName -Destination $target -Recurse -Force
    }
    foreach ($rel in $required) {
        if (!(Test-Path -LiteralPath (Join-Path $target $rel) -PathType Leaf)) {
            throw "Import verification failed: $rel"
        }
    }
    & git -C $gitRoot add -- Windows
    if ($LASTEXITCODE -ne 0) { throw "git add failed" }
    & git -C $gitRoot commit -m "Import Windows v7.5.8 full source, installer, firmware with original paths"
    if ($LASTEXITCODE -ne 0) { throw "git commit failed" }
    if (!$SkipPush) {
        & git -C $gitRoot push origin feature/pn15-grip
        if ($LASTEXITCODE -ne 0) { throw "git push failed; changes remain committed locally" }
    }
    Write-Host "Windows package imported to $target"
} finally {
    if (Test-Path -LiteralPath $tmp) { Remove-Item -LiteralPath $tmp -Force -Recurse }
}
