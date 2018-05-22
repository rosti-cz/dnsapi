dep:
	go get

vet: dep
	DNSAPI_PRIMARY_NAME_SERVER=ns1.rosti.cz \
	DNSAPI_NAME_SERVERS=ns1.rosti.cz,ns2.rosti.cz \
	DNSAPI_ABUSE_EMAIL=cx@initd.cz \
	go vet

test: dep
	DNSAPI_PRIMARY_NAME_SERVER=ns1.rosti.cz \
	DNSAPI_NAME_SERVERS=ns1.rosti.cz,ns2.rosti.cz \
	DNSAPI_ABUSE_EMAIL=cx@initd.cz \
	go test -run '' -v

cover: dep
	DNSAPI_PRIMARY_NAME_SERVER=ns1.rosti.cz \
	DNSAPI_NAME_SERVERS=ns1.rosti.cz,ns2.rosti.cz \
	DNSAPI_ABUSE_EMAIL=cx@initd.cz \
	go test -run '' -cover -coverprofile cover.out
	go tool cover -html=cover.out
