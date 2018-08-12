# DNS API

Simple DNS API for managing Bind server from remote localtion.

We use this as a micro service for our web administration. 

## Endpoints

The API covers two record types. One is for zones and the other one for records. Record is always grouped by zone.
Everytime you do a change and want to write it into NS servers call commit endpoint.

### Zones

    GET    /zones/
    
List of zones.

--

    GET    /zones/:zone_id

Returns data for *zone_id*.

--

    POST   /zones/

    JSON body:
        domain: domain name
        tags: tags separated by comma
        abuse_email: email for SOA record

Adds new zone.

--

    DELETE /zones/:zone_id

Deletes zone with *zone_id*.

--

    PUT    /zones/:zone_id
    
    JSON body:
        tags: tags separated by comma
        abuse_email: email for SOA record

Updates the zone *zone_id*.

--

    PUT    /zones/:zone_id/commit

Writes changes into the DNS servers.

### Records
    
    GET    /zones/:zone_id/records/
    
Returns list of records for *zone_id*.

--
    
    GET    /zones/:zone_id/records/:record_id
    
Return data for *record_id*.

--

    POST   /zones/:zone_id/records/

    JSON body:
        name: name of the record, ex. rosti.cz. or @
        ttl: time to live, ex. 3600
        type: record type, ex. A, AAAA, CNAME, ...
        prio: priority, only for MX
        value: value of the record

Adds a new record.

--

    DELETE /zones/:zone_id/records/:record_id
    
Deletes the record with *record_id*.

--

    PUT    /zones/:zone_id/records/:record_id
    
    JSON body:
        name: name of the record, ex. rosti.cz. or @
        ttl: time to live, ex. 3600
        type: record type, ex. A, AAAA, CNAME, ...
        prio: priority, only for MX
        value: value of the record

Updates the *record_id* with given data.