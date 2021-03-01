@echo off
SET BUILD_DIR=.\\pi
SET DEPLOY_DIR=/home/pi/Projects/discord-coc-bot

call pi-stop

echo Creating directory in pi
ssh pi@192.168.1.5 "mkdir -p %DEPLOY_DIR%" 
echo Transfering files from %BUILD_DIR% to %DEPLOY_DIR%
scp %BUILD_DIR%\* pi@192.168.1.5:%DEPLOY_DIR%
ssh pi@192.168.1.5 "chmod 700 -f %DEPLOY_DIR%/*"

call pi-start
