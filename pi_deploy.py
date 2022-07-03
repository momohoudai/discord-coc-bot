import subprocess
import os

user = "pi@192.168.0.69"
build_path = os.getcwd() + "/pi/"
deploy_path_in_pi = "/home/pi/Projects/discord-coc-bot"
service_name = "discord-coc-bot"

files = ["discord-coc-bot.exe", "TOKEN", "data.json"]
#if first time, use this instead. TODO: use a flag!
#files = ["discord-coc-bot.exe", "TOKEN", "data.json", "db"]

print("Stopping service")
subprocess.run(["ssh", user, "sudo service", service_name, "stop"])

print("Transfering files")
subprocess.run(["ssh", user, "mkdir -p", deploy_path_in_pi])
for e in files:
    subprocess.run(["scp", build_path + e, user + ":" + deploy_path_in_pi])

subprocess.run(["ssh", user, "chmod 700 -f", deploy_path_in_pi + "/*"])


print("Starting service")
subprocess.run(["ssh", user, "sudo service", service_name, "start"])


