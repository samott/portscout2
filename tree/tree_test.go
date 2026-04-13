package tree

import (
	"testing"
)

func TestParsePortConfig(t *testing.T) {
	result, err := parsePortConfig("site:http://www.x.com skipb:false limitw:1,odd skipv:1.1,1.9")

	if err != nil {
		t.Fatal("Port config parse error")
	}

	if result.IndexSite.String() != "http://www.x.com" {
		t.Fatal("Incorrect indexSite value")
	}

	if result.SkipBeta {
		t.Fatal("Incorrect skipBeta value")
	}

	if result.LimitWhich != 1 || result.LimitEven {
		t.Fatal("Incorrect limitWhich values")
	}

	if len(result.SkipVersions) != 2 {
		t.Fatal("Incorrect skipVersions count")
	}

	if result.SkipVersions[0] != "1.1" {
		t.Fatal("Incorrect skipVersions parse (entry 0)")
	}

	if result.SkipVersions[1] != "1.9" {
		t.Fatal("Incorrect skipVersions parse (entry 1)")
	}

	if result.Ignore {
		t.Fatal("Incorrect ignore value")
	}

	result2, err := parsePortConfig("skipb:true limitw:2,even skipv:2.234 ignore:true")

	if err != nil {
		t.Fatal("Port config parse error")
	}

	if result2.IndexSite != nil {
		t.Fatal("Incorrect indexSite value")
	}

	if !result2.SkipBeta {
		t.Fatal("Incorrect skipBeta value")
	}

	if result2.LimitWhich != 2 || !result2.LimitEven {
		t.Fatal("Incorrect limitWhich values")
	}

	if len(result2.SkipVersions) != 1 {
		t.Fatal("Incorrect skipVersions count")
	}

	if result2.SkipVersions[0] != "2.234" {
		t.Fatal("Incorrect skipVersions parse (entry 0)")
	}

	if !result2.Ignore {
		t.Fatal("Incorrect ignore value")
	}

	result3, err := parsePortConfig("limit:^[0-9.]+$")

	if !result3.LimitVer.MatchString("1.234") {
		t.Fatal("Regexp positive match failed");
	}

	if result3.LimitVer.MatchString("a1.234") {
		t.Fatal("Regexp negative match failed");
	}
}
