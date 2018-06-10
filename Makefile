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
	$GOPATH/bin/golinty

build: test
	go build -ldflags '-w' -o dnsapi

deploy: build
	scp dnsapi rosti-ns1:/opt/dnsapi_waiting_to_deploy
	ssh rosti-ns1 systemctl stop dnsapi
	ssh rosti-ns1 mv /opt/dnsapi_waiting_to_deploy /opt/dnsapi
	git rev-parse HEAD | ssh rosti-ns1 tee /opt/deployed
	ssh rosti-ns1 systemctl start dnsapi
