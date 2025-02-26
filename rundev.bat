@echo off

set PKG_PATH=github.com/jaeiya/billbank/lib/utils

for /f "delims=" %%a in ('git rev-parse HEAD') do set COMMIT_SHA=%%a
go run -ldflags "-X %PKG_PATH%.commitSha=%COMMIT_SHA% -X %PKG_PATH%.codeName=alpha" %*
