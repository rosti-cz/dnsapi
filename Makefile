dep:
	go get

vet: dep
	go vet

test: dep
	go test -run '' -v

cover: dep
	go test -run '' -cover -coverprofile cover.out
	go tool cover -html=cover.out
