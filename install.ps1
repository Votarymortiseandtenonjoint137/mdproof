#Requires -Version 5.1
<#
.SYNOPSIS
    Install mdproof on Windows
.DESCRIPTION
    Downloads and installs the latest mdproof release from GitHub
.EXAMPLE
    irm https://raw.githubusercontent.com/runkids/mdproof/main/install.ps1 | iex
#>

$ErrorActionPreference = "Stop"

# PowerShell 5.1 defaults to TLS 1.0/1.1; GitHub requires TLS 1.2+
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

$Repo = "runkids/mdproof"
$BinaryName = "mdproof"

function Write-Info { param($Message) Write-Host $Message -ForegroundColor Green }
function Write-Warn { param($Message) Write-Host $Message -ForegroundColor Yellow }
function Write-Err { param($Message) Write-Host $Message -ForegroundColor Red; exit 1 }

# Detect architecture
function Get-Arch {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        default { Write-Err "Unsupported architecture: $arch" }
    }
}

# Get latest version using redirect (avoids API rate limit)
function Get-LatestVersion {
    try {
        $response = Invoke-WebRequest -Uri "https://github.com/$Repo/releases/latest" `
            -MaximumRedirection 0 -UseBasicParsing -ErrorAction SilentlyContinue
    } catch {
        # PowerShell 5.1 throws on 3xx redirects; extract from the exception
        $response = $_.Exception.Response
    }

    $location = ""
    if ($response -and $response.Headers -and $response.Headers["Location"]) {
        $location = $response.Headers["Location"]
    }

    if ($location -match "/tag/([^/\s]+)") {
        return $Matches[1]
    }

    # Fallback to API if redirect fails
    try {
        $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
        return $release.tag_name
    } catch {
        Write-Err "Failed to get latest version. Please check your internet connection."
    }
}

# Get install directory
function Get-InstallDir {
    $dir = "$env:LOCALAPPDATA\Programs\mdproof"
    if (-not (Test-Path $dir)) {
        New-Item -ItemType Directory -Path $dir -Force | Out-Null
    }
    return $dir
}

# Add to PATH if not already present
function Add-ToPath {
    param($Dir)

    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($currentPath -notlike "*$Dir*") {
        Write-Info "Adding $Dir to PATH..."
        [Environment]::SetEnvironmentVariable("Path", "$currentPath;$Dir", "User")
        $env:Path = "$env:Path;$Dir"
        return $true
    }
    return $false
}

function Install-Mdproof {
    Write-Info "Installing mdproof..."
    Write-Host ""

    $arch = Get-Arch
    $version = Get-LatestVersion
    $installDir = Get-InstallDir

    # Release artifact: mdproof-v1.2.3-windows-amd64.zip
    $url = "https://github.com/$Repo/releases/download/$version/${BinaryName}-${version}-windows-${arch}.zip"

    Write-Info "Downloading mdproof $version for windows/$arch..."

    # Create temp directory
    $tempDir = Join-Path $env:TEMP "mdproof-install-$(Get-Random)"
    New-Item -ItemType Directory -Path $tempDir -Force | Out-Null

    try {
        $zipPath = Join-Path $tempDir "mdproof.zip"

        # Download
        Invoke-WebRequest -Uri $url -OutFile $zipPath -UseBasicParsing

        # Extract
        Expand-Archive -Path $zipPath -DestinationPath $tempDir -Force

        # Find the binary (named mdproof-vX.Y.Z-windows-amd64.exe)
        $exePath = Get-ChildItem -Path $tempDir -Filter "$BinaryName-*.exe" | Select-Object -First 1
        if (-not $exePath) {
            Write-Err "Binary not found in archive"
        }

        $destPath = Join-Path $installDir "$BinaryName.exe"
        Move-Item -Path $exePath.FullName -Destination $destPath -Force

        # Add to PATH
        $pathAdded = Add-ToPath -Dir $installDir

        Write-Host ""
        Write-Info "Successfully installed mdproof to $destPath"
        Write-Host ""

        # Show version
        & $destPath --version

        Write-Host ""
        if ($pathAdded) {
            Write-Warn "PATH updated. Restart your terminal for changes to take effect."
            Write-Host ""
        }

        Write-Info "Get started:"
        Write-Info "  mdproof --help"
        Write-Host ""
        Write-Warn "Note: Sandbox mode and jq assertions require WSL2 or Docker Desktop."
        Write-Warn "  https://learn.microsoft.com/en-us/windows/wsl/install"

    } finally {
        # Cleanup
        Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}

Install-Mdproof
