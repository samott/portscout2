package tree

import (
	"testing"
)

func TestParsePortConfig(t *testing.T) {
	result, err := parsePortConfig("site:http://www.x.com skipb:false limitw:1,odd")

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
		t.Fatal("Incorrect limitw values")
	}
}
