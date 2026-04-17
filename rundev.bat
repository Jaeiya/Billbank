@echo off

set PKG_PATH=github.com/jaeiya/billbank/internal/utils
set CMD_PATH=.\cmd\billbank\

for /f "delims=" %%a in ('git rev-parse HEAD') do set COMMIT_SHA=%%a
go run -ldflags "-X %PKG_PATH%.commitSha=%COMMIT_SHA% -X %PKG_PATH%.codeName=alpha" %CMD_PATH%
