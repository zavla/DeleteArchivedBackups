git.exe rev-list --pretty="format:%%h %%ai %%ae" -1 HEAD > ./gitCommit
for /f "usebackq tokens=*" %%i in (gitCommit) DO set gg=%%i


go build -v -ldflags="-X 'main.gitCommit=%gg%'" 


