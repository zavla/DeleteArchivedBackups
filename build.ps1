git.exe rev-list --pretty=medium -1 HEAD > ./gitCommit
$gg = Get-Content ./gitCommit

go build -v -ldflags="-X 'main.gitCommit=$gg'" 

