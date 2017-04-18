@echo off
setlocal
cls

for /f "delims=" %%a in ('git branch ^| grep "^*" ^| sed -e "s/* \(release\/\)\?\(.*\)/\2/"') do @set TAG=%%a
echo TAG  = %TAG%

echo.
echo ### Building version %TAG%...
echo.

rm -fr dist
mkdir dist

set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=0
go build -a --ldflags='-s' -o dist/proxy -v

set IMG=163759002681.dkr.ecr.eu-west-1.amazonaws.com/off-sync/platform-proxy:%TAG%

docker build -t %IMG% . && docker push %IMG%