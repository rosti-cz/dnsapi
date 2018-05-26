dep:
	go get

vet: dep
	go vet

test: dep
	go test -run '' -v

cover: dep
	go test -run '' -cover -coverprofile cover.out
	go tool cover -html=cover.out

cyclo:
	go get github.com/fzipp/gocyclo
	$GOPATH/bin/gocyclo  -over 15 *.go

lint:
	go get golang.org/x/lint/golint
	$GOPATH/bin/golint
