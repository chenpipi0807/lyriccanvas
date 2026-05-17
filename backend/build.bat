@echo on
set GOPROXY=https://goproxy.cn,direct
cd /d "%~dp0"
REM 生成 Windows exe 图标资源
rsrc -ico ..\frontend\public\favicon.ico
go build -v -o lyriccanvas.exe . 2>&1
