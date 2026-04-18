param(
    [switch]$System,
    [string]$InstallDir = "",
    [string]$BinaryName = "glagent.exe"
)

$ErrorActionPreference = "Stop"

$root = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
Push-Location $root

try {
    $args = @("run", ".", "setup")
    if ($System) {
        $args += "--system"
    }
    if ($InstallDir -ne "") {
        $args += "--install-dir"
        $args += $InstallDir
    }
    if ($BinaryName -ne "") {
        $args += "--binary-name"
        $args += $BinaryName
    }

    & go @args
}
finally {
    Pop-Location
}
