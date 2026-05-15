param(
    [string]$Version = $(if ($env:BOLTZ_API_VERSION) { $env:BOLTZ_API_VERSION } else { "latest" }),
    [string]$InstallDir = $env:BOLTZ_API_INSTALL_DIR,
    [string]$InstallBaseUrl = $(if ($env:BOLTZ_API_INSTALL_BASE_URL) { $env:BOLTZ_API_INSTALL_BASE_URL } else { "https://install.boltz.bio/boltz-api" }),
    [int]$ReleaseRetries = $(if ($env:BOLTZ_API_RELEASE_RETRIES) { [int]$env:BOLTZ_API_RELEASE_RETRIES } else { 12 }),
    [int]$ReleaseRetryDelaySeconds = $(if ($env:BOLTZ_API_RELEASE_RETRY_DELAY) { [int]$env:BOLTZ_API_RELEASE_RETRY_DELAY } else { 10 })
)

$ErrorActionPreference = "Stop"

function Fail($Message) {
    Write-Error $Message
    exit 1
}

if ($ReleaseRetries -lt 0) {
    Fail "BOLTZ_API_RELEASE_RETRIES must be a non-negative integer"
}

if ($ReleaseRetryDelaySeconds -lt 0) {
    Fail "BOLTZ_API_RELEASE_RETRY_DELAY must be a non-negative integer"
}

function Get-ConfigFilePath {
    $base = [Environment]::GetFolderPath("ApplicationData")
    if (-not $base) {
        $base = Join-Path $HOME ".config"
    }
    Join-Path (Join-Path $base "boltz-compute") "config.yaml"
}

function Get-YamlScalar($Content, $Key) {
    foreach ($line in $Content) {
        if ($line -match "^\s*$([regex]::Escape($Key)):\s*['""]?([^'""]*)['""]?\s*$") {
            return $Matches[1]
        }
    }
    return ""
}

function Warn-ExistingConfig {
    $configFile = Get-ConfigFilePath
    if (-not (Test-Path $configFile)) {
        return
    }

    $content = Get-Content -Path $configFile -ErrorAction SilentlyContinue
    $configIssuer = Get-YamlScalar $content "issuer_url"
    $configClient = Get-YamlScalar $content "client_id"

    if ($configIssuer -and $configIssuer -ne "https://lab.boltz.bio") {
        Write-Warning "Existing boltz-api config at $configFile sets auth issuer to $configIssuer."
        Write-Warning "Run 'boltz-api config show' to inspect it or 'boltz-api config reset' to remove non-secret local config."
    }
    if ($configClient -and $configClient -ne "boltz-cli") {
        Write-Warning "Existing boltz-api config at $configFile sets auth client ID to $configClient."
        Write-Warning "Run 'boltz-api config show' to inspect it or 'boltz-api config reset' to remove non-secret local config."
    }
}

Warn-ExistingConfig

function ConvertTo-InstallArch($Architecture) {
    if ($null -eq $Architecture) {
        return ""
    }

    $value = $Architecture.ToString().Trim()
    if (-not $value) {
        return ""
    }

    switch -Regex ($value) {
        "^(X64|AMD64|x86_64)$" { return "amd64" }
        "^(X86|x86|I386|I686|386)$" { return "386" }
        "^(Arm64|ARM64|AARCH64)$" { return "arm64" }
        default { return "" }
    }
}

function Get-InstallArch {
    $candidates = @()
    try {
        $candidates += [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture
    } catch {
        # Windows PowerShell on older .NET Frameworks may not expose RuntimeInformation.
    }
    $candidates += $env:PROCESSOR_ARCHITEW6432
    $candidates += $env:PROCESSOR_ARCHITECTURE

    foreach ($candidate in $candidates) {
        $normalized = ConvertTo-InstallArch $candidate
        if ($normalized) {
            return $normalized
        }
    }

    $candidateList = @(
        $candidates |
            Where-Object { $null -ne $_ -and $_.ToString().Trim() } |
            ForEach-Object { $_.ToString().Trim() }
    ) -join ", "
    if (-not $candidateList) {
        $candidateList = "unknown"
    }

    Fail "Unsupported CPU architecture: $candidateList"
}

$arch = Get-InstallArch

if ($Version -eq "latest") {
    $InstallBaseUrl = $InstallBaseUrl.TrimEnd("/")
    $releaseUrl = "$InstallBaseUrl/latest.json"
    $allowReleaseFallback = $true
} else {
    if ($Version.StartsWith("v")) {
        $tag = $Version
    } else {
        $tag = "v$Version"
    }
    $InstallBaseUrl = $InstallBaseUrl.TrimEnd("/")
    $releaseUrl = "$InstallBaseUrl/releases/$tag/release.json"
    $allowReleaseFallback = $false
}

$script:retry = 0
while ($true) {
    try {
        $releaseResponse = Invoke-RestMethod -Uri $releaseUrl
    } catch {
        Fail "Could not fetch boltz-api release metadata from $releaseUrl"
    }

    $releaseCandidates = @($releaseResponse)
    $latestTag = $releaseCandidates[0].tag_name
    if (-not $latestTag) {
        Fail "Could not determine the boltz-api release tag"
    }

    $release = $null
    $asset = $null
    foreach ($candidate in $releaseCandidates) {
        $candidateAsset = $candidate.assets |
            Where-Object { $_.name -match "^boltz-api_.*_windows_${arch}\.zip$" } |
            Select-Object -First 1
        if ($candidateAsset) {
            $release = $candidate
            $asset = $candidateAsset
            if ($asset.browser_download_url -match "/releases/(?:download/)?([^/]+)/") {
                $release.tag_name = $Matches[1]
            }
            break
        }
        if (-not $allowReleaseFallback) {
            $release = $candidate
            break
        }
    }

    if ($asset) {
        if ($allowReleaseFallback -and $release.tag_name -ne $latestTag) {
            Write-Warning "Latest boltz-api release $latestTag has no windows/$arch asset yet; installing $($release.tag_name) instead."
        }
        break
    }

    if ($script:retry -ge $ReleaseRetries) {
        Fail "No boltz-api release asset found for windows/$arch in $latestTag after $ReleaseRetries retries"
    }

    $script:retry += 1
    Write-Warning "No boltz-api release asset found for windows/$arch in $latestTag; retrying in ${ReleaseRetryDelaySeconds}s ($script:retry/$ReleaseRetries)"
    Start-Sleep -Seconds $ReleaseRetryDelaySeconds
}

$versionNumber = $release.tag_name -replace "^v", ""

if (-not $InstallDir) {
    $existing = Get-Command boltz-api -ErrorAction SilentlyContinue
    if ($existing -and $existing.Source) {
        $existingBinary = $existing.Source
        $InstallDir = Split-Path -Parent $existing.Source
    } elseif ($env:LOCALAPPDATA) {
        $existingBinary = $null
        $InstallDir = Join-Path $env:LOCALAPPDATA "Programs\Boltz\bin"
    } else {
        $existingBinary = $null
        $InstallDir = Join-Path $HOME ".local\bin"
    }
} else {
    $existingBinary = Join-Path $InstallDir "boltz-api.exe"
}

if ($existingBinary -and (Test-Path $existingBinary)) {
    $currentOutput = & $existingBinary --version 2>$null
    if ($currentOutput -match "([0-9]+\.[0-9][^ ]*)") {
        if ($Matches[1] -eq $versionNumber) {
            Write-Host "boltz-api $($release.tag_name) is already installed at $existingBinary"
            exit 0
        }
    }
}

$tmp = Join-Path ([System.IO.Path]::GetTempPath()) ("boltz-api-" + [System.Guid]::NewGuid())
New-Item -ItemType Directory -Path $tmp | Out-Null

try {
    $archive = Join-Path $tmp "boltz-api.zip"
    Invoke-WebRequest -Uri $asset.browser_download_url -OutFile $archive
    Expand-Archive -Path $archive -DestinationPath $tmp -Force

    $binary = Get-ChildItem -Path $tmp -Filter "boltz-api.exe" -Recurse | Select-Object -First 1
    if (-not $binary) {
        Fail "Downloaded archive did not contain boltz-api.exe"
    }

    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    Copy-Item -Path $binary.FullName -Destination (Join-Path $InstallDir "boltz-api.exe") -Force
} finally {
    Remove-Item -Path $tmp -Recurse -Force -ErrorAction SilentlyContinue
}

Write-Host "Installed boltz-api $($release.tag_name) to $(Join-Path $InstallDir "boltz-api.exe")"

$pathEntries = $env:Path -split ";"
if ($pathEntries -notcontains $InstallDir) {
    Write-Warning "Add $InstallDir to PATH to run boltz-api without the full path."
}
