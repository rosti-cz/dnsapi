package main

import (
	"testing"
	"fmt"
	"crypto/sha256"
)

func ExampleZone() {
	// A valid zone
	zone, errs := NewZone(TEST_DOMAIN, []string{"test_tag_1", "test_tag_2"}, TEST_ABUSE_EMAIL)
	if len(errs) != 0 {
		panic(errs)
	}
	zone.SetNewSerial()

	zone.AddRecord("@", 300, "A", 0, "1.2.3.4")
	zone.AddRecord("@", 300, "AAAA", 0, "2001::2")
	zone.AddRecord("www", 300, "CNAME", 0, "@")
	zone.AddRecord("@", 300, "MX", 10, "mail.rosti.cz.")
	zone.AddRecord("@", 300, "TXT", 0, "igeeweofeiroomoogokieghaithohthaechoocherohveehiebawuyeixeecoveegoeyohfachainauquaeceetipheivubohmoegheizeelaiquanaokooquiedokaidurahveehoshazaseveitheiyitachudiishaeghaexoovachacaijuyiedeochojingafeexusuquaingeiboovachahlaechahcashoophairohthaghobahjaixieboteixameimohmaedahriebaekoshohpeecueyaoseeveibavaithohquaevoalohreingewiesaijiojiehielahzaelohpechuohiefaeyaetiegengahgatheefaipeimeeviedimibohmoyajefahghaaraehieyiepameegheathaechielixahbeidohyaitionahjaenoshikahbahyaebeachahxalaeghuloochaekuthaiquaedoo")
	zone.Serial = "201805300001"
	renderedZone := zone.Render()

	h := sha256.New()
	h.Write([]byte(renderedZone))
	fmt.Printf("%x", h.Sum(nil))
	// Output: 29100887d2a05dddcdfd49391ebed51fd487fc6df565ee2b70302d7d1bf2bc78
}

func ExampleZone_RenderPrimary() {
	zone, errs := NewZone(TEST_DOMAIN, []string{"test_tag_1", "test_tag_2"}, TEST_ABUSE_EMAIL)
	if len(errs) != 0 {
		panic(errs)
	}
	zone.SetNewSerial()

	zone.AddRecord("@", 300, "A", 0, "1.2.3.4")

	fmt.Println(zone.RenderPrimary())
	// Output:
	// zone "ohphiuhi.txt" IN {
	//         type master;
	//         file "zones/ohphiuhi.txt.zone";
	//         allow-query { any; };
	//         allow-transfer { 5.6.7.8 };
	//         notify yes;
	// };
}

func ExampleZone_RenderSecondary() {
	zone, errs := NewZone(TEST_DOMAIN, []string{"test_tag_1", "test_tag_2"}, TEST_ABUSE_EMAIL)
	if len(errs) != 0 {
		panic(errs)
	}
	zone.SetNewSerial()

	zone.AddRecord("@", 300, "A", 0, "1.2.3.4")

	fmt.Println(zone.RenderSecondary())
	// Output:
	// zone "ohphiuhi.txt" IN {
	//     type slave;
	//     file "zones/ohphiuhi.txt.zone";
	//     allow-query { any; };
	//     masters { 1.2.3.4 };
	// };
}

func TestValidZone(t *testing.T) {
	// A valid zone with config's email
	var zone Zone
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
		Serial: "200601020001",
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

	zone, errs := NewZone(TEST_DOMAIN, []string{"test_tag_1", "test_tag_2"}, TEST_ABUSE_EMAIL)
	if len(errs) != 0 {
		t.Error(errs)
	}

	zone.AddRecord("@", 300, "A", 0, "1.2.3.a") // Invalid IPv4
	zone.AddRecord("@", 300, "AAAA", 0, "2001::g") // Invalid IPv6
	zone.AddRecord("www", 300, "A", 0, "1.2.3.4") // Valid A record
	zone.AddRecord("www", 300, "CNAME", 0, "@") // same name as existing a record
	zone.AddRecord("abc", 300, "CNAME", 0, "||||") // invalid value in CNAME
	zone.AddRecord("@", 300, "MX", 0, "mail.rosti.cz.") // invalid prio
	zone.AddRecord("@", 1, "MX", 10, "mail.rosti.cz.") // invalid TTL
	zone.AddRecord("@", 300, "UNKNOWN", 0, "mail.rosti.cz.") // invalid record type
	zone.AddRecord("@", 300, "TXT", 0, "\"igeeweofeiroomoogokieghaithohthaechoocherohveehiebawuyeixeecoveegoeyohfachainauquaeceetipheivubohmoegheizeelaiquanaokooquiedokaidurahveehoshazaseveitheiyitachudiishaeghaexoovachacaijuyiedeochojingafeexusuquaingeiboovachahlaechahcashoophairohthaghobahjaixieboteixameimohmaedahriebaekoshohpeecueyaoseeveibavaithohquaevoalohreingewiesaijiojiehielahzaelohpechuohiefaeyaetiegengahgatheefaipeimeeviedimibohmoyajefahghaaraehieyiepameegheathaechielixahbeidohyaitionahjaenoshikahbahyaebeachahxalaeghuloochaekuthaiquaedoo")

	errs = zone.Validate()
	// TODO: check exact errors
	if len(errs) != 8 {
		t.Error("Not right amount of errors were generated", errs)
	}
}
