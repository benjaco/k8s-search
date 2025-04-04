update:
  go get -u
  go mod tidy -v

compile:
  go build -o secret-search.exe main.go
  GOOS=darwin GOARCH=arm64 go build -o secret-search main.go