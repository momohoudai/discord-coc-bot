@echo off
SET BUILD_DIR=.\\pi
SET GOARCH=arm
SET GOOS=linux
SET CGO_ENABLED=0

echo Creating folder: %BUILD_DIR%
IF NOT EXIST "%BUILD_DIR%" mkdir %BUILD_DIR%
go build -o %BUILD_DIR%\\discord-coc-bot.exe
echo Built
