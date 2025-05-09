@ECHO OFF

del xg.exe
go.exe build -ldflags="-s -w" ./cmd/xg/
