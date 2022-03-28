
@REM build client
set GOARCH=amd64
set GOOS=windows
go build -a -o bin\client.exe  .\client\main.go




@REM build macos server
set GOARCH=amd64
set GOOS=darwin
go build -a -o bin\darwin_server .\server\server.go
