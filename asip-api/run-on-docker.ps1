<#
.SYNOPSIS
    Build and run the as-ip API Docker stack locally or on a remote host via SSH.

.DESCRIPTION
    Uses docker-compose.yml from the project root. When --ssh-string is omitted, Docker on
    localhost is used. When --ssh-string is set, the image is built locally, exported,
    transferred to the remote host, loaded there, and started without rebuilding on the server.
    When --delete-volume=yes, existing containers and named volumes are removed before start.

.EXAMPLE
    .\run-on-docker.ps1

.EXAMPLE
    .\run-on-docker.ps1 --delete-volume=yes

.EXAMPLE
    .\run-on-docker.ps1 --ssh-string=myvps

.EXAMPLE
    .\run-on-docker.ps1 --ssh-string=production --delete-volume=no
#>
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$Script:ProjectRoot = $PSScriptRoot
$Script:ComposeFile = 'docker-compose.yml'
$Script:DockerNetwork = 'asip-net'
$Script:RemoteProjectPath = '/opt/as-ip'
$Script:ImageName = 'asip-api:latest'
$Script:ImageArchiveName = 'asip-api.tar'
$Script:LocalDeployDir = Join-Path $Script:ProjectRoot '.deploy'

function Show-RunOnDockerHelp {
    Write-Host @'
as-ip Docker run - build and start the API stack

Usage:
  .\run-on-docker.ps1 [--ssh-string=<alias>] [--delete-volume=<no|yes>] [--help]

Arguments:
  --ssh-string=<alias>        SSH config alias for remote Docker (e.g. myvps)
                              The script prepends "ssh" when connecting; do not include "ssh"
                              in the value. Builds locally, transfers the image, then starts
                              the stack remotely. When omitted, localhost Docker is used.
  --delete-volume=<no|yes>    Remove named volumes before starting (default: no)
  --help, -h                  Show this help message and exit

Examples:
  .\run-on-docker.ps1
  .\run-on-docker.ps1 --help
  .\run-on-docker.ps1 --delete-volume=yes
  .\run-on-docker.ps1 --ssh-string=myvps
  .\run-on-docker.ps1 --ssh-string=production --delete-volume=no

Local API URL:  http://localhost:3000/api/v1/health
'@ -ForegroundColor Cyan
}

function ConvertTo-RunArguments {
    param([string[]]$RawArguments)

    $parsed = @{
        ssh_string     = $null
        delete_volume  = 'no'
        help           = $false
    }

    foreach ($argument in $RawArguments) {
        if ($argument -match '^--(?<name>[\w-]+)(?:=(?<value>.*))?$') {
            $key = ($Matches['name'] -replace '-', '_').ToLowerInvariant()
            $value = if ($Matches.ContainsKey('value')) { $Matches['value'] } else { 'true' }

            switch ($key) {
                'help' { $parsed['help'] = $true }
                'ssh_string' { $parsed['ssh_string'] = $value.Trim() }
                'delete_volume' { $parsed['delete_volume'] = $value.Trim().ToLowerInvariant() }
                default { throw "Unknown argument: --$($Matches['name']). Run with --help." }
            }
        }
        elseif ($argument -match '^(-h|-help|--help|-\?|/\?)$') {
            $parsed['help'] = $true
        }
        else {
            throw "Unknown argument: $argument. Run with --help."
        }
    }

    if ($parsed['delete_volume'] -notin @('no', 'yes', 'false', 'true', '0', '1')) {
        throw "Invalid --delete-volume value '$($parsed['delete_volume'])'. Allowed: no, yes."
    }

    return $parsed
}

function Test-DeleteVolumeEnabled {
    param([string]$Value)

    return $Value -in @('yes', 'true', '1')
}

function Resolve-SshAlias {
    param([string]$SshString)

    $alias = $SshString.Trim()

    if ($alias -match '^(?i)ssh(\s|$)') {
        throw 'Invalid --ssh-string value. Pass only the SSH config alias (e.g. --ssh-string=myvps). Do not include "ssh".'
    }

    if ([string]::IsNullOrWhiteSpace($alias)) {
        throw 'Invalid --ssh-string value. Example: --ssh-string=myvps'
    }

    return $alias
}

function Write-RunStep {
    param(
        [int]$Step,
        [int]$Total,
        [string]$Message
    )

    $percent = [math]::Round(($Step / $Total) * 100)
    Write-Progress -Activity 'as-ip Docker run' -Status $Message -PercentComplete $percent
    Write-Host ("[{0}/{1}] {2}" -f $Step, $Total, $Message) -ForegroundColor Yellow
}

function Test-DockerCliAvailable {
    param([string]$CommandPrefix = '')

    $checkCommand = if ($CommandPrefix) { "$CommandPrefix docker version" } else { 'docker version' }
    Invoke-Expression $checkCommand | Out-Null

    if ($LASTEXITCODE -ne 0) {
        throw 'Docker CLI is not available or not running.'
    }
}

function Invoke-RemoteShell {
    param(
        [string]$SshAlias,
        [string]$Command,
        [string]$WorkingDirectory = $null
    )

    $remoteCommand = if ($WorkingDirectory) { "cd '$WorkingDirectory' && $Command" } else { $Command }
    & ssh $SshAlias $remoteCommand

    if ($LASTEXITCODE -ne 0) {
        throw "Remote command failed (exit $LASTEXITCODE): $remoteCommand"
    }
}

function Build-LocalDockerImage {
    Push-Location $Script:ProjectRoot
    try {
        docker compose -f $Script:ComposeFile build
        if ($LASTEXITCODE -ne 0) {
            throw 'Local docker compose build failed.'
        }
    }
    finally {
        Pop-Location
    }
}

function Export-LocalDockerImage {
    param([string]$ArchivePath)

    $parentDirectory = Split-Path -Parent $ArchivePath
    if (-not (Test-Path -LiteralPath $parentDirectory)) {
        New-Item -ItemType Directory -Path $parentDirectory -Force | Out-Null
    }

    if (Test-Path -LiteralPath $ArchivePath) {
        Remove-Item -LiteralPath $ArchivePath -Force
    }

    docker save $Script:ImageName -o $ArchivePath
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to export image '$($Script:ImageName)'."
    }
}

function Sync-ComposeToRemote {
    param([string]$SshAlias)

    & ssh $SshAlias "mkdir -p '$Script:RemoteProjectPath'"
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to create remote directory: $Script:RemoteProjectPath"
    }

    $composePath = Join-Path $Script:ProjectRoot $Script:ComposeFile
    $remoteDestination = '{0}:{1}/' -f $SshAlias, $Script:RemoteProjectPath
    & scp -o StrictHostKeyChecking=accept-new $composePath $remoteDestination
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to copy '$Script:ComposeFile' to remote host."
    }
}

function Transfer-ImageToRemote {
    param(
        [string]$SshAlias,
        [string]$ArchivePath
    )

    $remoteArchivePath = '{0}:{1}/{2}' -f $SshAlias, $Script:RemoteProjectPath, $Script:ImageArchiveName
    & scp -o StrictHostKeyChecking=accept-new $ArchivePath $remoteArchivePath
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to transfer image archive to remote host."
    }
}

function Import-RemoteDockerImage {
    param([string]$SshAlias)

    $remoteArchivePath = '{0}/{1}' -f $Script:RemoteProjectPath, $Script:ImageArchiveName
    Invoke-RemoteShell -SshAlias $SshAlias -Command "docker load -i '$remoteArchivePath' && rm -f '$remoteArchivePath'"
}

function Ensure-DockerNetwork {
    param([string]$CommandPrefix = '')

    if ($CommandPrefix) {
        $createCommand = "$CommandPrefix docker network inspect '$Script:DockerNetwork' >/dev/null 2>&1 || $CommandPrefix docker network create '$Script:DockerNetwork'"
        Invoke-Expression $createCommand | Out-Null
        if ($LASTEXITCODE -ne 0) {
            throw "Failed to ensure Docker network '$Script:DockerNetwork'."
        }
        return
    }

    $existingNetworks = docker network ls --format '{{.Name}}'
    if ($LASTEXITCODE -ne 0) {
        throw 'Failed to list Docker networks. Is Docker running?'
    }
    if ($existingNetworks -notcontains $Script:DockerNetwork) {
        docker network create $Script:DockerNetwork | Out-Null
        if ($LASTEXITCODE -ne 0) {
            throw "Failed to create Docker network '$Script:DockerNetwork'."
        }
    }
}

function Get-ComposeEnvironmentPrefix {
    param([bool]$BehindProxy)

    $prefix = "DOCKER_NETWORK='$($Script:DockerNetwork)' "
    if ($BehindProxy) {
        $prefix += "API_PUBLISH_PORT='' "
    }
    return $prefix
}

function Invoke-DockerComposeRun {
    param(
        [string]$CommandPrefix = '',
        [string]$WorkingDirectory = $Script:ProjectRoot,
        [bool]$DeleteVolume = $false,
        [bool]$Build = $true,
        [bool]$BehindProxy = $false
    )

    Push-Location $WorkingDirectory
    try {
        $envPrefix = Get-ComposeEnvironmentPrefix -BehindProxy:$BehindProxy
        $composeDown = if ($CommandPrefix) {
            "$CommandPrefix ${envPrefix}docker compose -f $Script:ComposeFile down"
        }
        else {
            "${envPrefix}docker compose -f $Script:ComposeFile down"
        }

        if ($DeleteVolume) {
            $composeDown += ' -v'
        }

        Invoke-Expression $composeDown | Out-Null
        if ($LASTEXITCODE -ne 0) {
            Write-Host 'Compose down skipped or partial (stack may not exist yet).' -ForegroundColor DarkYellow
        }

        $composeUp = if ($CommandPrefix) {
            "$CommandPrefix ${envPrefix}docker compose -f $Script:ComposeFile up -d"
        }
        else {
            "${envPrefix}docker compose -f $Script:ComposeFile up -d"
        }

        if ($Build) {
            $composeUp += ' --build'
        }

        Invoke-Expression $composeUp
        if ($LASTEXITCODE -ne 0) {
            throw 'docker compose up failed.'
        }
    }
    finally {
        Pop-Location
    }
}

function Invoke-LocalDockerRun {
    param([bool]$DeleteVolume)

    Write-RunStep -Step 1 -Total 4 -Message 'Checking local Docker'
    Test-DockerCliAvailable

    Write-RunStep -Step 2 -Total 4 -Message "Ensuring Docker network '$($Script:DockerNetwork)'"
    Ensure-DockerNetwork

    Write-RunStep -Step 3 -Total 4 -Message $(if ($DeleteVolume) { 'Stopping stack and removing volumes' } else { 'Stopping stack (keeping volumes)' })
    Invoke-DockerComposeRun -DeleteVolume:$DeleteVolume -Build

    Write-RunStep -Step 4 -Total 4 -Message 'Stack started'
    Write-Progress -Activity 'as-ip Docker run' -Completed -Status 'Done'
    Write-Host ''
    Write-Host 'Run complete. API available at http://localhost:3000/api/v1/health' -ForegroundColor Green
}

function Invoke-RemoteDockerRun {
    param(
        [string]$SshAlias,
        [bool]$DeleteVolume
    )

    $archivePath = Join-Path $Script:LocalDeployDir $Script:ImageArchiveName

    Write-RunStep -Step 1 -Total 7 -Message 'Checking local Docker'
    Test-DockerCliAvailable

    Write-RunStep -Step 2 -Total 7 -Message "Building image locally ($($Script:ImageName))"
    Build-LocalDockerImage

    Write-RunStep -Step 3 -Total 7 -Message 'Exporting image archive'
    Export-LocalDockerImage -ArchivePath $archivePath

    Write-RunStep -Step 4 -Total 7 -Message "Checking Docker on $SshAlias"
    Test-DockerCliAvailable -CommandPrefix "ssh $SshAlias"

    Write-RunStep -Step 5 -Total 7 -Message "Transferring image and compose file to $SshAlias"
    Sync-ComposeToRemote -SshAlias $SshAlias
    Transfer-ImageToRemote -SshAlias $SshAlias -ArchivePath $archivePath
    Import-RemoteDockerImage -SshAlias $SshAlias

    Write-RunStep -Step 6 -Total 7 -Message "Ensuring Docker network '$($Script:DockerNetwork)' on $SshAlias"
    Ensure-DockerNetwork -CommandPrefix "ssh $SshAlias"

    Write-RunStep -Step 7 -Total 7 -Message $(if ($DeleteVolume) { 'Stopping remote stack and removing volumes' } else { 'Stopping remote stack (keeping volumes)' })
    $envPrefix = Get-ComposeEnvironmentPrefix -BehindProxy
    $composeDown = "${envPrefix}docker compose -f $Script:ComposeFile down"
    if ($DeleteVolume) {
        $composeDown += ' -v'
    }

    try {
        Invoke-RemoteShell -SshAlias $SshAlias -Command $composeDown -WorkingDirectory $Script:RemoteProjectPath
    }
    catch {
        Write-Host 'Remote compose down skipped or partial (stack may not exist yet).' -ForegroundColor DarkYellow
    }

    Invoke-RemoteShell -SshAlias $SshAlias -Command "${envPrefix}docker compose -f $Script:ComposeFile up -d" -WorkingDirectory $Script:RemoteProjectPath

    if (Test-Path -LiteralPath $archivePath) {
        Remove-Item -LiteralPath $archivePath -Force
    }

    Write-Progress -Activity 'as-ip Docker run' -Completed -Status 'Done'
    Write-Host ''
    Write-Host ("Run complete on {0}. Image built locally and deployed without remote build." -f $SshAlias) -ForegroundColor Green
}

$argumentMap = ConvertTo-RunArguments -RawArguments $args
if ($argumentMap['help']) {
    Show-RunOnDockerHelp
    exit 0
}

$deleteVolume = Test-DeleteVolumeEnabled -Value $argumentMap['delete_volume']
$sshString = $argumentMap['ssh_string']

try {
    if ([string]::IsNullOrWhiteSpace($sshString)) {
        Write-Host 'Target: localhost Docker' -ForegroundColor Cyan
        Invoke-LocalDockerRun -DeleteVolume:$deleteVolume
    }
    else {
        $sshAlias = Resolve-SshAlias -SshString $sshString
        Write-Host ("Target: remote Docker via ssh {0} (local build + image transfer)" -f $sshAlias) -ForegroundColor Cyan
        Invoke-RemoteDockerRun -SshAlias $sshAlias -DeleteVolume:$deleteVolume
    }
}
catch {
    Write-Progress -Activity 'as-ip Docker run' -Completed -Status 'Failed'
    Write-Host ''
    Write-Host $_.Exception.Message -ForegroundColor Red
    Write-Host ''
    Show-RunOnDockerHelp
    exit 1
}
