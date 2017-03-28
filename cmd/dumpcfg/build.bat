@echo off
setlocal
cls

pushd "." & for %%i in (.) do set REPO=%%~ni
echo REPO = %REPO%

for /f "delims=" %%a in ('git branch ^| grep "^*" ^| sed -e "s/* \(release\/\)\?\(.*\)/\2/"') do @set TAG=%%a
echo TAG  = %TAG%

rm -fr dist
mkdir dist

set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=0
go build -a --ldflags='-s' -o dist/%REPO% -v

set IMG=163759002681.dkr.ecr.eu-west-1.amazonaws.com/off-sync/platform-proxy-dumpcfg:%TAG%
echo %IMG%

docker build -t %IMG% . && docker push %IMG%
