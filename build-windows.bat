@echo off
setlocal

go build -ldflags="-H=windowsgui -s -w" -o CodexConfigAssistant.exe .
if errorlevel 1 exit /b %errorlevel%

echo.
echo Build completed: CodexConfigAssistant.exe
