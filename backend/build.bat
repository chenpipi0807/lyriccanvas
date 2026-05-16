@echo on
set GOPROXY=https://goproxy.cn,direct
cd /d "%~dp0"
go build -v -o lyriccanvas.exe . 2>&1
