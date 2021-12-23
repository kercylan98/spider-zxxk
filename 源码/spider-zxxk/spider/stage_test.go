package spider

import "testing"

func TestSpider_GetStages(t *testing.T) {
	spider := New()

	stages, err := spider.GetStages()
	if err != nil {
		t.Fatal(err)
	}

	for _, stage := range stages {
		t.Log(stage)
	}
}
