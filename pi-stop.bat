@echo off

echo Stopping service
ssh pi@192.168.1.5 "sudo supervisorctl stop discord-coc-bot"
