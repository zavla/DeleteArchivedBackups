git log --pretty="%h %ae %cd" -1 > ./gitCommit

$gg = Get-Content ./gitCommit

go build -v -ldflags="-X 'main.gitCommit=$gg'" 

