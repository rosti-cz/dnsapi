package main

import (
	"testing"
	"fmt"
	"crypto/sha256"
	"time"
)

func ExampleZone() {
	// A valid zone
	zone, errs := NewZone("B-" + TEST_DOMAIN, []string{"test_tag_1", "test_tag_2"}, TEST_ABUSE_EMAIL)
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
	// Output: 94cd266e6c539311b2786706e10e0172be30e8fc8c49db78de38517c886c4fb7
}

func ExampleZone_RenderPrimary() {
	zone, errs := NewZone("C-" + TEST_DOMAIN, []string{"test_tag_1", "test_tag_2"}, TEST_ABUSE_EMAIL)
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
	zone, errs := NewZone("D-" + TEST_DOMAIN, []string{"test_tag_1", "test_tag_2"}, TEST_ABUSE_EMAIL)
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
	if zone.Serial != today + "01" {
		t.Error("Got " + zone.Serial + ", expected "+ today +"01")
	}

	zone.SetNewSerial()
	if zone.Serial != today + "02" {
		t.Error("Got " + zone.Serial + ", expected "+ today +"02")
	}

	zone.SetNewSerial()
	if zone.Serial != today + "03" {
		t.Error("Got " + zone.Serial + ", expected "+ today +"03")
	}

	zone.SetNewSerial()
	if zone.Serial != today + "04" {
		t.Error("Got " + zone.Serial + ", expected "+ today +"04")
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
		Domain: "F-" + TEST_DOMAIN,
		Serial: "2006010201",
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

	zone, errs := NewZone("G-" + TEST_DOMAIN, []string{"test_tag_1", "test_tag_2"}, TEST_ABUSE_EMAIL)
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
