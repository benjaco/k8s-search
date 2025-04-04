update:
  go get -u
  go mod tidy -v

compile:
  go build -o secret_search.exe main.go
  GOOS=darwin GOARCH=arm64 go build -o secret_search main.go