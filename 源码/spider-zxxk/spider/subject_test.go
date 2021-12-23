package spider

import "testing"

func TestSpider_GetSubjects(t *testing.T) {
	spider := New()

	subjects, err := spider.GetSubjects(`4`)
	if err != nil {
		t.Fatal(err)
	}

	for _, subject := range subjects {
		t.Log(subject)
	}
}
//https://yw.zxxk.com/h/books/
//https://yw.zxxk.com/books