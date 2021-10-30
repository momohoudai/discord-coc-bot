@echo off

SET USER=pi@192.168.0.69
SET BUILD_DIR=.\\pi
SET DEPLOY_DIR=/home/pi/Projects/discord-coc-bot
SET SERVICE_NAME=discord-coc-bot

echo Stopping service
ssh %USER% "sudo supervisorctl stop %SERVICE_NAME%"

echo Creating directory in pi
ssh %USER% "mkdir -p %DEPLOY_DIR%" 
echo Transfering files from %BUILD_DIR% to %DEPLOY_DIR%
scp %BUILD_DIR%\* %USER%:%DEPLOY_DIR%
ssh %USER% "chmod 700 -f %DEPLOY_DIR%/*"

echo Starting service
ssh %USER% "sudo supervisorctl start %SERVICE_NAME%"