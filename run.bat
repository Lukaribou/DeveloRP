@ECHO OFF
CLS 

IF "%1%"=="linux" (
    SET GOOS=linux
    SET GOARCH=arm
    SET GOARM=7
)

go build

IF "%1%"=="linux" (
    SET GOOS=windows
    SET GOARCH=amd64
    ECHO Build reussi pour Linux ARM7
) ELSE (DeveloRP.exe)
