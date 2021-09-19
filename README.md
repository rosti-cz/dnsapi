# DNS API

Simple DNS API for managing Bind server from remote localtion.

We use this as a micro service for our web administration. 

Currently is runs on the master server and changes the secondary one via SSH.

Make sure you use either private network or webserver with TLS in front of the service before run in production.

## TODO

* deletetion of unexisted zones causes "Internal Server Error" (I am not sure it still happens)
* Make master and slave modes

## Installation

Installation:

```
wget -O /usr/local/bin/dnsapi https://github.com/rosti-cz/dnsapi/releases/download/v1.0/dnsapi-1.0-linux-amd64
chmod +x /usr/local/bin/dnsapi
```

Check MD5 sum:

```
wget -O - -q https://github.com/rosti-cz/dnsapi/releases/download/v1.0/md5sums | grep `md5sum /usr/local/bin/dnsapi`
```

Copy this service file into */etc/systemd/system/dnsapi.service* and update its configuration values:

```
[Unit]
Description=DNSAPI micro service
After=network.target

[Service]
PIDFile=/run/dnsapi.pid
WorkingDirectory=/etc/bind

Environment=DNSAPI_PRIMARY_NAME_SERVER_IP=1.2.3.4
Environment=DNSAPI_SECONDARYNAMESERVERIPS=2.3.4.5
Environment=DNSAPI_PRIMARY_NAME_SERVER=ns1.examepl.com
Environment=DNSAPI_NAME_SERVERS=ns1.example.com,ns2.example.com
Environment=DNSAPI_ABUSE_EMAIL=info@example.com
Environment=DNSAPI_DATABASE_PATH=/var/lib/dnsapi/db.sqlite
Environment=DNSAPI_SSHKEY=/root/.ssh/id_rsa
Environment=DNSAPI_SSHUSER=root
Environment=DNSAPI_APITOKEN=ABCDE
Environment=DNSAPI_PORT=80

ExecStart=/usr/local/bin/dnsapi
ExecStop=/bin/kill -s TERM $MAINPID
PrivateTmp=false

[Install]
WantedBy=multi-user.target
```

Enable and start:

```
systemctl start dnsapi
systemctl enable dnsapi
```

## Endpoints

The API covers two record types. One is for zones and the other one for records. Record is always grouped by zone.
Everytime you do a change and want to write it into NS servers call commit endpoint.

### Zones

    GET    /zones/
    
List of zones.

---

    GET    /zones/:zone_id

Returns data for *zone_id*.

---

    POST   /zones/

    JSON body:
        domain: domain name
        tags: tags separated by comma
        abuse_email: email for SOA record

Adds new zone.

---

    DELETE /zones/:zone_id

Deletes zone with *zone_id*.

---

    PUT    /zones/:zone_id
    
    JSON body:
        tags: tags separated by comma
        abuse_email: email for SOA record

Updates the zone *zone_id*.

---

    PUT    /zones/:zone_id/commit

Writes changes into the DNS servers.

### Records
    
    GET    /zones/:zone_id/records/
    
Returns list of records for *zone_id*.

---
    
    GET    /zones/:zone_id/records/:record_id
    
Return data for *record_id*.

---

    POST   /zones/:zone_id/records/

    JSON body:
        name: name of the record, ex. rosti.cz. or @
        ttl: time to live, ex. 3600
        type: record type, ex. A, AAAA, CNAME, ...
        prio: priority, only for MX
        value: value of the record

Adds a new record.

---

    DELETE /zones/:zone_id/records/:record_id
    
Deletes the record with *record_id*.

---

    PUT    /zones/:zone_id/records/:record_id
    
    JSON body:
        name: name of the record, ex. rosti.cz. or @
        ttl: time to live, ex. 3600
        type: record type, ex. A, AAAA, CNAME, ...
        prio: priority, only for MX
        value: value of the record

Updates the *record_id* with given data.
