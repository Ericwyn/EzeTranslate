@echo off
setlocal

echo start build EzeTranslate
echo.

go build EzeTranslate.go
if errorlevel 1 exit /b %errorlevel%

for /f %%i in ('EzeTranslate.exe -v') do set VER_CODE=%%i
echo Version Code: %VER_CODE%

for /f "tokens=2 delims==" %%a in ('wmic OS Get localdatetime /value') do set datetime=%%a
set CURRENT_TIME=%datetime:~2,12%
echo Current Time: %CURRENT_TIME%

set TARGET_DIR=.\build-target\%VER_CODE%_windows_%CURRENT_TIME%
if not exist "%TARGET_DIR%" mkdir "%TARGET_DIR%"

copy ".\EzeTranslate.exe" "%TARGET_DIR%\EzeTranslate.exe" >nul

echo Build and packaging completed successfully.
echo.
echo target build path: %TARGET_DIR%
