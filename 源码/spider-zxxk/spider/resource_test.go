package spider

import (
	"fmt"
	"testing"
)

func TestSpider_GetResource(t *testing.T) {
	spider := New()



	resources, err := spider.GetResource("https://sx.zxxk.com/p/books")
	if err != nil {
		panic(err)
	}

	s, err := spider.GetResourceChildren(resources[0])
	if err != nil {
		panic(err)
	}

	ss, err := spider.GetResourceChildren(s[5])
	if err != nil {
		panic(err)
	}

	sss, err := spider.GetResourceChildren(ss[7])
	if err != nil {
		panic(err)
	}

	ssss, err := spider.GetResourceChildren(sss[12])
	if err != nil {
		panic(err)
	}

	sssss, err := spider.GetResourceChildren(ssss[0])
	if err != nil {
		panic(err)
	}

	for _, resource := range sssss {
		fmt.Println(resource)
	}

	//for _, resource := range resources {
	//	resources, err := spider.GetResourceChildren(resource)
	//	if err != nil {
	//		log.Fatalln(err)
	//	}
	//	for _, r := range resources {
	//		fmt.Println(r)
	//		rrs, err := spider.GetResourceChildren(r)
	//		if err != nil {
	//			panic(rrs)
	//		}
	//		for _, r := range rrs {
	//			fmt.Println(r)
	//		}
	//	}
	//	return
	//}
}
