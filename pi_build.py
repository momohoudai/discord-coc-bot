import subprocess
import os

build_path = "pi"

if not os.path.exists(build_path):
    os.mkdir(build_path)

os.environ["GOARCH"] = 'arm'
os.environ["GOOS"] = 'linux'
os.environ["CGO_ENABLED"] = '0'

subprocess.run(["go", "build", "-o", build_path+"/discord-coc-bot.exe"])
print("Built");