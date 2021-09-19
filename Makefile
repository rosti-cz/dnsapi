VERSION=1.0

.PHONY: dep
dep:
	go mod tidy

.PHONY: vet
vet: dep
	go vet

.PHONY: test
test: dep
	go test -run '' -v

.PHONY: cover
cover: dep
	go test -run '' -cover -coverprofile cover.out
	go tool cover -html=cover.out

.PHONY: cyclo
cyclo:
	go get github.com/fzipp/gocyclo
	$GOPATH/bin/gocyclo  -over 15 *.go

.PHONY: lint
lint:
	go get golang.org/x/lint/golint
	$GOPATH/bin/golinty

.PHONY: clean
clean:
	rm -f bin/*

.PHONY: init
init:
	mkdir -p ./bin

.PHONY: build
build: test clean init linux-amd64 linux-arm linux-arm64
	md5sum bin/dnsapi-* > bin/md5sums
	sha256sum bin/dnsapi-* > bin/sha256sums

.PHONY: linux-amd64
linux-amd64: init dep clean
	# linux amd64
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/dnsapi-${VERSION}-linux-amd64 *.go

.PHONY: linux-arm
linux-arm: init dep clean
	# linux arm
	env CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -o ./bin/dnsapi-${VERSION}-linux-arm *.go

.PHONY: linux-arm64
linux-arm64: init dep clean
	# linux arm64
	env CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o ./bin/dnsapi-${VERSION}-linux-arm64 *.go

.PHONY: deploy
deploy:
	scp dnsapi rosti-ns1:/opt/dnsapi_waiting_to_deploy
	ssh rosti-ns1 systemctl stop dnsapi
	ssh rosti-ns1 mv /opt/dnsapi_waiting_to_deploy /opt/dnsapi
	git rev-parse HEAD | ssh rosti-ns1 tee /opt/deployed
	ssh rosti-ns1 systemctl start dnsapi
