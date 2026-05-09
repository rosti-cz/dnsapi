package dnsapi

import (
	"crypto/sha256"
	"fmt"
	"testing"
	"time"
)

func TestRecordRenderSRVQuotesValue(t *testing.T) {
	record := Record{
		Name:  "_sip._tcp",
		TTL:   300,
		Type:  "SRV",
		Value: "10 5 5060 sip.example.com.",
	}

	rendered := record.Render()
	expected := "_sip._tcp    300s    SRV      10 5 5060 sip.example.com."

	if rendered != expected {
		t.Fatalf("unexpected SRV rendering: got %q, expected %q", rendered, expected)
	}
}

func ExampleZone() {
	// A valid zone
	zone, errs := NewZone("B-"+TEST_DOMAIN, []string{"test_tag_1", "test_tag_2"}, TEST_ABUSE_EMAIL, "testowner")
	if len(errs) != 0 {
		panic(errs)
	}
	zone.SetNewSerial()

	zone.AddRecord("@", 300, "A", 0, "1.2.3.4")
	zone.AddRecord("@", 300, "AAAA", 0, "2001::2")
	zone.AddRecord("www", 300, "CNAME", 0, "@")
	zone.AddRecord("@", 300, "MX", 10, "mail.rosti.cz.")
	zone.AddRecord("@", 300, "TXT", 0, "igeeweofeiroomoogokieghaithohthaechoocherohveehiebawuyeixeecoveegoeyohfachainauquaeceetipheivubohmoegheizeelaiquanaokooquiedokaidurahveehoshazaseveitheiyitachudiishaeghaexoovachacaijuyiedeochojingafeexusuquaingeiboovachahlaechahcashoophairohthaghobahjaixieboteixameimohmaedahriebaekoshohpeecueyaoseeveibavaithohquaevoalohreingewiesaijiojiehielahzaelohpechuohiefaeyaetiegengahgatheefaipeimeeviedimibohmoyajefahghaaraehieyiepameegheathaechielixahbeidohyaitionahjaenoshikahbahyaebeachahxalaeghuloochaekuthaiquaedoo")
	zone.Serial = "2018053001"
	renderedZone := zone.Render()

	h := sha256.New()
	h.Write([]byte(renderedZone))
	fmt.Printf("%x", h.Sum(nil))
	// Output: 635d2be2a44773e301a0938357b868dbfe5282b2e516e28d21c4b0e41d642333
}

func ExampleZone_RenderPrimary() {
	zone, errs := NewZone("C-"+TEST_DOMAIN, []string{"test_tag_1", "test_tag_2"}, TEST_ABUSE_EMAIL, "testowner")
	if len(errs) != 0 {
		panic(errs)
	}
	zone.SetNewSerial()

	zone.AddRecord("@", 300, "A", 0, "1.2.3.4")

	fmt.Println(zone.RenderPrimary())
	// Output:
	// zone "c-ohphiuhi.txt" IN {
	//         type master;
	//         masterfile-format text;
	//         file "c-ohphiuhi.txt.zone";
	//         allow-query { any; };
	//         allow-transfer { 5.6.7.8; };
	//         notify yes;
	// };
}

func ExampleZone_RenderSecondary() {
	zone, errs := NewZone("D-"+TEST_DOMAIN, []string{"test_tag_1", "test_tag_2"}, TEST_ABUSE_EMAIL, "testowner")
	if len(errs) != 0 {
		panic(errs)
	}
	zone.SetNewSerial()

	zone.AddRecord("@", 300, "A", 0, "1.2.3.4")

	fmt.Println(zone.RenderSecondary())
	// Output:
	// zone "d-ohphiuhi.txt" IN {
	//     type slave;
	//     masterfile-format text;
	//     file "d-ohphiuhi.txt.zone";
	//     allow-query { any; };
	//     masters { 1.2.3.4; };
	// };
}

func TestZone_SetNewSerial(t *testing.T) {
	var zone Zone

	today := time.Now().UTC().Format("20060102")

	zone.SetNewSerial()
	if zone.Serial != today+"01" {
		t.Error("Got " + zone.Serial + ", expected " + today + "01")
	}

	zone.SetNewSerial()
	if zone.Serial != today+"02" {
		t.Error("Got " + zone.Serial + ", expected " + today + "02")
	}

	zone.SetNewSerial()
	if zone.Serial != today+"03" {
		t.Error("Got " + zone.Serial + ", expected " + today + "03")
	}

	zone.SetNewSerial()
	if zone.Serial != today+"04" {
		t.Error("Got " + zone.Serial + ", expected " + today + "04")
	}
}

func TestValidZone(t *testing.T) {
	// A valid zone with config's email
	var zone = Zone{
		Domain: "E-" + TEST_DOMAIN,
	}
	zone.AddRecord("@", 300, "A", 0, "1.2.3.4")
	zone.AddRecord("@", 300, "AAAA", 0, "2001::2")
	zone.AddRecord("www", 300, "CNAME", 0, "@")
	zone.AddRecord("@", 300, "MX", 10, "mail.rosti.cz.")
	zone.AddRecord("@", 300, "TXT", 0, "eigeeweofeiroomoogokieghaithohthaechoocherohveehiebawuyeixeecoveegoeyohfachainauquaeceetipheivubohmoegheizeelaiquanaokooquiedokaidurahveehoshazaseveitheiyitachudiishaeghaexoovachacaijuyiedeochojingafeexusuquaingeiboovachahlaechahcashoophairohthaghobahjaixieboteixameimohmaedahriebaekoshohpeecueyaoseeveibavaithohquaevoalohreingewiesaijiojiehielahzaelohpechuohiefaeyaetiegengahgatheefaipeimeeviedimibohmoyajefahghaaraehieyiepameegheathaechielixahbeidohyaitionahjaenoshikahbahyaebeachahxalaeghuloochaekuthaiquaedoo")

	errs := zone.Validate()
	if len(errs) > 0 {
		t.Error(errs)
	}
	zone.Render()

	// A valid zone with config's email
	zone = Zone{
		Domain:     "F-" + TEST_DOMAIN,
		Serial:     "2006010201",
		AbuseEmail: "cx@initd.cz",
	}
	zone.AddRecord("@", 300, "A", 0, "1.2.3.4")
	zone.AddRecord("@", 300, "AAAA", 0, "2001::2")
	zone.AddRecord("www", 300, "CNAME", 0, "@")
	zone.AddRecord("@", 300, "MX", 10, "mail.rosti.cz.")
	zone.AddRecord("@", 300, "TXT", 0, "eigeeweofeiroomoogokieghaithohthaechoocherohveehiebawuyeixeecoveegoeyohfachainauquaeceetipheivubohmoegheizeelaiquanaokooquiedokaidurahveehoshazaseveitheiyitachudiishaeghaexoovachacaijuyiedeochojingafeexusuquaingeiboovachahlaechahcashoophairohthaghobahjaixieboteixameimohmaedahriebaekoshohpeecueyaoseeveibavaithohquaevoalohreingewiesaijiojiehielahzaelohpechuohiefaeyaetiegengahgatheefaipeimeeviedimibohmoyajefahghaaraehieyiepameegheathaechielixahbeidohyaitionahjaenoshikahbahyaebeachahxalaeghuloochaekuthaiquaedoo")

	errs = zone.Validate()
	if len(errs) > 0 {
		t.Error(errs)
	}
	zone.Render()
}

func TestInvalidZone(t *testing.T) {
	// A valid zone
	var zone *Zone

	zone, errs := NewZone("G-"+TEST_DOMAIN, []string{"test_tag_1", "test_tag_2"}, TEST_ABUSE_EMAIL, "testowner")
	if len(errs) != 0 {
		t.Error(errs)
	}

	zone.AddRecord("@", 300, "A", 0, "1.2.3.a")              // Invalid IPv4
	zone.AddRecord("@", 300, "AAAA", 0, "2001::g")           // Invalid IPv6
	zone.AddRecord("www", 300, "A", 0, "1.2.3.4")            // Valid A record
	zone.AddRecord("www", 300, "CNAME", 0, "@")              // same name as existing a record
	zone.AddRecord("abc", 300, "CNAME", 0, "||||")           // invalid value in CNAME
	zone.AddRecord("@", 300, "MX", -1, "mail.rosti.cz.")     // invalid prio
	zone.AddRecord("@", 1, "MX", 10, "mail.rosti.cz.")       // invalid TTL
	zone.AddRecord("@", 300, "UNKNOWN", 0, "mail.rosti.cz.") // invalid record type
	zone.AddRecord("@", 300, "TXT", 0, "\"igeeweofeiroomoogokieghaithohthaechoocherohveehiebawuyeixeecoveegoeyohfachainauquaeceetipheivubohmoegheizeelaiquanaokooquiedokaidurahveehoshazaseveitheiyitachudiishaeghaexoovachacaijuyiedeochojingafeexusuquaingeiboovachahlaechahcashoophairohthaghobahjaixieboteixameimohmaedahriebaekoshohpeecueyaoseeveibavaithohquaevoalohreingewiesaijiojiehielahzaelohpechuohiefaeyaetiegengahgatheefaipeimeeviedimibohmoyajefahghaaraehieyiepameegheathaechielixahbeidohyaitionahjaenoshikahbahyaebeachahxalaeghuloochaekuthaiquaedoo")

	errs = zone.Validate()
	// TODO: check exact errors
	if len(errs) != 8 {
		t.Error("Not right amount of errors were generated", errs)
	}
}

func TestValidRecordTypes(t *testing.T) {
	tests := []struct {
		name   string
		record Record
	}{
		{"NS", Record{Name: "@", TTL: 300, Type: "NS", Value: "ns1.example.com."}},
		{"NS no trailing dot", Record{Name: "sub", TTL: 300, Type: "NS", Value: "ns1.example.com"}},
		{"CAA issue", Record{Name: "@", TTL: 300, Type: "CAA", Value: `0 issue "letsencrypt.org"`}},
		{"CAA issuewild", Record{Name: "@", TTL: 300, Type: "CAA", Value: `128 issuewild "pki.goog"`}},
		{"CAA iodef", Record{Name: "@", TTL: 300, Type: "CAA", Value: `0 iodef "mailto:admin@example.com"`}},
		{"SRV", Record{Name: "_sip._tcp", TTL: 300, Type: "SRV", Value: "10 20 5060 sip.example.com."}},
		{"MX prio 0", Record{Name: "@", TTL: 300, Type: "MX", Prio: 0, Value: "mail.example.com."}},
		{"MX prio 65535", Record{Name: "@", TTL: 300, Type: "MX", Prio: 65535, Value: "mail.example.com."}},
		{"MX prio 200", Record{Name: "@", TTL: 300, Type: "MX", Prio: 200, Value: "mail.example.com."}},
		{"CNAME with underscore", Record{Name: "sub", TTL: 300, Type: "CNAME", Value: "_dmarc.example.com."}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.record.Validate(); err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
		})
	}
}

func TestInvalidRecordTypes(t *testing.T) {
	tests := []struct {
		name   string
		record Record
	}{
		{"NS bad hostname", Record{Name: "@", TTL: 300, Type: "NS", Value: "not valid!"}},
		{"CAA bad tag", Record{Name: "@", TTL: 300, Type: "CAA", Value: `0 badtag "value"`}},
		{"CAA missing quotes", Record{Name: "@", TTL: 300, Type: "CAA", Value: "0 issue letsencrypt.org"}},
		{"SRV bad format", Record{Name: "@", TTL: 300, Type: "SRV", Value: "not valid srv"}},
		{"SRV missing port", Record{Name: "@", TTL: 300, Type: "SRV", Value: "10 20 sip.example.com."}},
		{"MX prio -1", Record{Name: "@", TTL: 300, Type: "MX", Prio: -1, Value: "mail.example.com."}},
		{"MX prio 65536", Record{Name: "@", TTL: 300, Type: "MX", Prio: 65536, Value: "mail.example.com."}},
		{"MX bad hostname", Record{Name: "@", TTL: 300, Type: "MX", Prio: 10, Value: "not valid!"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.record.Validate(); err == nil {
				t.Errorf("expected validation error, got none")
			}
		})
	}
}

func TestInvalid2Zone(t *testing.T) {
	// A valid zone
	var zone *Zone

	zone, errs := NewZone("G-"+TEST_DOMAIN+"2", []string{"test_tag_1", "test_tag_2"}, TEST_ABUSE_EMAIL, "testowner")
	if len(errs) != 0 {
		t.Error(errs)
	}

	zone.AddRecord("@", 300, "CNAME", 0, "a.b.") // Invalid IPv4

	errs = zone.Validate()
	// TODO: check exact errors
	if len(errs) == 0 {
		t.Error("Expected validation errors, got none")
	}

	if len(errs) > 0 && errs[0].Error() != "CNAME record cannot be root record" {
		t.Error("Expected specific validation error, got:", errs[0])
	}
}
