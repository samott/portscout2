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
}
