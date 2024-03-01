@echo off
echo start build EzeTranslate
echo.

:: 编译EzeTranslate.go
go build EzeTranslate.go

:: 执行 ./EzeTranslate -v, 获取一个版本号 VER_CODE
for /f %%i in ('EzeTranslate -v') do set VER_CODE=%%i
echo Version Code: %VER_CODE%

:: 获取当前时间，格式为 yyMMddHHmmss
for /f "tokens=2 delims==" %%a in ('wmic OS Get localdatetime /value') do set datetime=%%a
set CURRENT_TIME=%datetime:~2,12%
echo Current Time: %CURRENT_TIME%

:: 创建文件夹 ./build-target/VER_CODE_{yyMMddHHmmss}
set TARGET_DIR=.\build-target\%VER_CODE%_windows_%CURRENT_TIME%
if not exist "%TARGET_DIR%" mkdir "%TARGET_DIR%"

:: 将 ./EzeTranslate, ./res-static/ ,./config.yaml 都复制到 ./build-target/VER_CODE_{yyMMddHHmmss}/
xcopy /E /I ".\EzeTranslate" "%TARGET_DIR%\EzeTranslate"
xcopy /E /I ".\res-static" "%TARGET_DIR%\res-static"
xcopy /E /I ".\config.yaml" "%TARGET_DIR%\config.yaml"

echo Build and packaging completed successfully.
echo.
echo target build path: %TARGET_DIR%

