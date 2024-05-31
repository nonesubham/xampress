@echo off
setlocal enabledelayedexpansion

rem List of OS/Arch combinations to build
set targets=linux/amd64 linux/386 linux/arm linux/arm64 windows/amd64 windows/386 darwin/amd64 darwin/arm64

rem Build the binary for each target
for %%t in (%targets%) do (
    for /f "tokens=1,2 delims=/" %%a in ("%%t") do (
        set "os=%%a"
        set "arch=%%b"
        set "output_name=hello-!os!-!arch!"

        if "!os!" == "windows" (
            set "output_name=!output_name!.exe"
        )

        echo Building for !os!/!arch!...
        set "GOOS=!os!"
        set "GOARCH=!arch!"
        call go build -o !output_name! main.go

        if errorlevel 1 (
            echo Failed to build for !os!/!arch!
        )
    )
)

echo Build completed.
endlocal
