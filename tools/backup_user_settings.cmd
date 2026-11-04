@echo off
setlocal EnableExtensions
set "SRC=%APPDATA%\PulseNoteDeck"
if not exist "%SRC%\config_unified.json" (
  echo PulseNote configuration was not found:
  echo %SRC%
  echo Check that PulseNote Deck has saved at least one profile.
  pause
  exit /b 1
)
for /f %%I in ('powershell -NoProfile -Command "Get-Date -Format yyyyMMdd_HHmmss"') do set "STAMP=%%I"
if not defined STAMP set "STAMP=backup"
set "DEST=%~dp0..\..\USER_SETTINGS_BACKUP_%STAMP%"
echo Source: "%SRC%"
echo Backup: "%DEST%"
robocopy "%SRC%" "%DEST%" /E /R:2 /W:1 /COPY:DAT /DCOPY:DAT
set "RESULT=%ERRORLEVEL%"
if %RESULT% GEQ 8 (
  echo Backup failed. Robocopy code %RESULT%.
  pause
  exit /b %RESULT%
)
echo.
echo Backup complete. Copy this USER_SETTINGS_BACKUP folder to safe storage.
echo Restore with the app fully closed, to %%APPDATA%%\PulseNoteDeck.
pause
exit /b 0
