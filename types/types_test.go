package types

import (
	"testing"
)

func TestUnmarshalTaggedLists(t *testing.T) {
	result := UnmarshalTaggedLists("http://x.com/:tag1,tag2 http://y.com/:tag1 http://z.com/:tag3 http://q.com http://fake.com:_cargo")

	if len(result) != 4 {
		t.Fatal("Incorrect tag count")
	}

	if _, ok := result[""]; !ok {
		t.Fatal("Missing tag: default tag")
	}

	if _, ok := result["tag1"]; !ok {
		t.Fatal("Missing tag: tag1")
	}

	if _, ok := result["tag2"]; !ok {
		t.Fatal("Missing tag: tag2")
	}

	if _, ok := result["tag3"]; !ok {
		t.Fatal("Missing tag: tag3")
	}

	if len(result["tag1"].Items) != 2 {
		t.Fatal("Incorrect item count for tag1")
	}

	if len(result["tag2"].Items) != 1 {
		t.Fatal("Incorrect item count for tag2")
	}

	if len(result["tag3"].Items) != 1 {
		t.Fatal("Incorrect item count for tag3")
	}
}

func TestMarshalTaggedLists(t *testing.T) {
	input := map[string]*TaggedList{
		"tag1": &TaggedList{
			Items: []string{
				"http://www.example.net",
				"http://www.example.com",
			},
		},
		"tag2": &TaggedList{
			Items: []string{
				"http://www.example.org",
			},
		},
		"tag3": &TaggedList{
			Items: []string{
				"http://www.example.org",
			},
		},
		"_cargo": &TaggedList{
			Items: []string{
				"http://www.example.org",
			},
		},
		"": &TaggedList{
			Items: []string{
				"http://www.microsoft.com",
			},
		},
	}

	output := MarshalTaggedLists(input)

	if output != "http://www.microsoft.com http://www.example.net:tag1 http://www.example.com:tag1 http://www.example.org:tag2 http://www.example.org:tag3" {
		t.Fatal("Incorrect marshaled tagged lists string:", output)
	}
}
