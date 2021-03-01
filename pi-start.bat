@echo off

echo Starting service
ssh pi@192.168.1.5 "sudo supervisorctl start discord-coc-bot"
