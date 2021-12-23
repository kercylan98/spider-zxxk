package main

import (
	"encoding/json"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"github.com/tealeg/xlsx"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"spider-zxxk/spider"
	"strings"
	"sync"
	"time"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

var templateFile, err = xlsx.OpenFile(`./教材目录导入模板.xlsx`)

var s = spider.New()
var pool *ants.Pool
var antErr error

func init() {
	pool, antErr = ants.NewPool(10)
	if antErr != nil {
		log.Fatal("线程池构建失败")
	}
}


func Save(path string, obj interface{}) error {
	p := strings.ReplaceAll(path, "*", "_")
	p = strings.ReplaceAll(p, " ", "_")

	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(p, data, 0666)
}

func Read(path string, obj interface{}) error {
	p := strings.ReplaceAll(path, "*", "_")
	p = strings.ReplaceAll(p, " ", "_")

	data, err := ioutil.ReadFile(strings.ReplaceAll(p, "*", "_"))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, obj)
}

func main() {
	checkErr(err)

	getStages: {}
	var stages []spider.Stage
	var err error
	if err = Read("./data/stages.json", &stages); err != nil {
		stages, err = s.GetStages()
		if err != nil {
			log.Println("获取学段信息失败，将在三秒后重试...", err)
			time.Sleep(3 * time.Second)
			goto getStages
		}else {
			if err = Save("./data/stages.json", &stages); err != nil {
				log.Println("学段信息保存失败", err)
			}
		}
	}


	var root = &Item{
		Name:   "根",
	}

	for _, stage := range stages {
		if exitSignal {
			continue
		}

		getSubjects: {}
		var path = fmt.Sprintf("./data/%s_subjects.json", stage.Name)
		var subjects []spider.Subject
		var err error
		if err = Read(path, &subjects); err != nil {
			subjects, err = s.GetSubjects(stage.LevelId)
			if err != nil {
				log.Println("获取学科信息失败，将在三秒后重试...", err)
				time.Sleep(3 * time.Second)
				goto getSubjects
			}else {
				if err = Save(path, &subjects); err != nil {
					log.Println("学科信息保存失败", err)
				}
			}
		}

		for _, subject := range subjects {
			if exitSignal {
				continue
			}

			getVersions: {}
			var path = fmt.Sprintf("./data/%s_%s_versions.json", stage.Name, subject.Name)
			var versions []spider.Resource
			var err error
			if err = Read(path, &versions); err != nil {
				versions, err = s.GetResource(subject.Url)
				if err != nil {
					log.Println(stage, subject, "获取教材目录信息失败，将在三秒后重试...", err)
					time.Sleep(3 * time.Second)
					goto getVersions
				}else {
					if err = Save(path, &versions); err != nil {
						log.Println("教材目录信息保存失败", err)
					}
				}
			}

			for i := 0; i < len(versions); i++ {
				version := versions[i]
				if version.Name == "教材" {

					// 装载锁
					if root.rootLock == nil {
						root.rootLock = new(sync.RWMutex)
					}

					if !exitSignal {
						antErr = pool.Submit(func() {
							// 多线程模式
							id := i
							Scan(id, root, version, stage.Name, subject.Name)
						})
						if antErr != nil {
							// 标准模式
							log.Println("线程池构建失败，启用标准模式")
							root.rootLock = nil
							Scan(i, root, version, stage.Name, subject.Name)
						}
					}

				}

			}
		}
	}

	retrySave: {}
	lineLock.Lock()
	defer lineLock.Unlock()
	if err := templateFile.Save("./result.xlsx"); err != nil {
		fmt.Println("保存Excel失败", err)
		time.Sleep(3 * time.Second)
		goto retrySave
	}
	log.Println("记录已存储...")

	log.Println("程序已退出")
}

var exitSignal = false
var exitSignalLock = new(sync.RWMutex)
var onceExit = new(sync.Once)
var saveLock = new(sync.RWMutex)
func Scan(threadId int, parent *Item, resource spider.Resource, stage string, subject string) (exit bool) {
	if data, err := ioutil.ReadFile("./save.signal"); err == nil {
		if strings.Contains(string(data), "yes") {
			saveLock.Lock()
		retrySave: {}
			lineLock.Lock()
			if err := templateFile.Save("./result.xlsx"); err != nil {
				fmt.Println("保存Excel失败", err)
				time.Sleep(3 * time.Second)
				goto retrySave
			}
			lineLock.Unlock()
			log.Println("记录已存储...")
			os.Remove("./save.signal")
			saveLock.Unlock()
		}
	}

	if data, err := ioutil.ReadFile("./stop.signal"); err == nil {
		if strings.Contains(string(data), "yes") {
			onceExit.Do(func() {
				exitSignalLock.Lock()
				exitSignal = true
				exitSignalLock.Unlock()
				log.Println("正在停止进程")
			})
		}
	}
	exitSignalLock.Lock()
	if exitSignal {
		log.Println(fmt.Sprintf("某个线程已退出，剩余 %d 个...", pool.Running()))
		exitSignalLock.Unlock()
		return true
	}
	exitSignalLock.Unlock()

	getChildren: {}
	var path = fmt.Sprintf("./data/%s_%s_%s_children.json", stage, subject, parent.PathTree())
	var children []spider.Resource
	var err error
	if err = Read(path, &children); err != nil {
		children, err = s.GetResourceChildren(resource)
		if err != nil {
			log.Println("获取资源失败，将在三秒后重试...", err)
			time.Sleep(3 * time.Second)
			goto getChildren
		}else {
			if err = Save(path, &children); err != nil {
				log.Println("资源保存失败", err)
			}
		}
	}

	for _, child := range children {
		if parent.rootLock != nil {
			parent.rootLock.Lock()
		}
		item := &Item{
			Name:  child.Name,
			Parent: parent,
			Level: parent.Level + 1,
			Stage: stage,
			Subject: subject,
		}
		parent.Items = append(parent.Items, item)
		item.Log(threadId, stage, subject)
		item.Insert()
		if parent.rootLock != nil {
			parent.rootLock.Unlock()
		}


		if Scan(threadId, item, child, stage, subject) {
			return true
		}

		// 清理内存
		if parent.rootLock != nil {
			parent.rootLock.Lock()
		}
		var newItems []*Item
		for i := 0; i < len(parent.Items); i++ {
			if parent.Items[i] != item {
				newItems = append(newItems, parent.Items[i])
			}
		}
		parent.Items = newItems
		runtime.GC()
		if parent.rootLock != nil {
			parent.rootLock.Unlock()
		}
	}

	return false
}

var lineLock = new(sync.RWMutex)
func addLine() *xlsx.Row {
	lineLock.Lock()
	defer lineLock.Unlock()
	row := templateFile.Sheets[0].AddRow()
	for i := 0; i < 12; i++ {
		row.AddCell()
	}
	row.Cells[11].SetString(time.Now().String())
	return row
}

type Item struct {
	Name string
	Items []*Item
	Parent *Item
	Level int
	Stage string
	Subject string
	fullGrade bool
	fullBook bool
	rootLock *sync.RWMutex
}

func (slf *Item) Log(threadId int, stage string, subject string) {
	var now = slf
	var str string
	for now != nil && now.Name != "根" {
		if str == "" {
			str = "[" + fmt.Sprint(now.Level) + "]" + now.Name
		}else {
			str = "[" + fmt.Sprint(now.Level) + "]" + now.Name + " > " + str
		}
		now = now.Parent
	}
	log.Println(fmt.Sprintf("[Thread: %d][%s][%s] > %s", threadId, stage, subject, str))
}

func (slf *Item) PathTree() string {
	var now = slf
	var str string
	for now != nil {
		n := strings.ReplaceAll(now.Name, "/", "_")
		n = strings.ReplaceAll(n, "*", "_")
		n = strings.ReplaceAll(n, "\\", "_")

		str = "(" + fmt.Sprint(now.Level) + ")" + n + "_" + str
		now = now.Parent
	}
	return str
}

func (slf *Item) FullGrade(row *xlsx.Row) {
	var grades []string
	switch slf.Stage {
	case "小学":
		grades = append(grades, "小学一年级", "小学二年级", "小学三年级", "小学四年级", "小学五年级", "小学六年级")
	case "初中":
		grades = append(grades, "预初", "初中一年级", "初中二年级", "初中三年级")
	case "高中":
		grades = append(grades, "高中一年级", "高中二年级", "高中三年级")
	}
	for i, grade := range grades {
		var newRow *xlsx.Row
		if i == 0 {
			newRow = row
		}else {
			newRow = addLine()
		}
		for i, cell := range newRow.Cells {
			cell.SetString(row.Cells[i].String())
		}
		newRow.Cells[1].SetString(grade)
	}
}

func (slf *Item) FullBook(row *xlsx.Row) {
	for i, book := range []string{"上册", "下册"} {
		var newRow *xlsx.Row
		if i == 0 {
			newRow = row
		}else {
			newRow = addLine()
		}
		for i, cell := range newRow.Cells {
			cell.SetString(row.Cells[i].String())
		}
		newRow.Cells[3].SetString(book)
	}
}

func (slf *Item) FullGradeAndBook(row *xlsx.Row) {
	var grades []string
	switch slf.Stage {
	case "小学":
		grades = append(grades, "小学一年级", "小学二年级", "小学三年级", "小学四年级", "小学五年级", "小学六年级")
	case "初中":
		grades = append(grades, "预初", "初中一年级", "初中二年级", "初中三年级")
	case "高中":
		grades = append(grades, "高中一年级", "高中二年级", "高中三年级")
	}
	for i, grade := range grades {
		var newRow *xlsx.Row
		if i == 0 {
			newRow = row
		}else {
			newRow = addLine()
		}
		for i, cell := range newRow.Cells {
			cell.SetString(row.Cells[i].String())
		}
		newRow.Cells[1].SetString(grade)
		slf.FullBook(newRow)
	}
}


func (slf *Item) Insert() {
	if slf.Level > 2 {
		row := addLine()

		now := slf
		for now.Level > 0 {
			switch now.Level {
			case 1:
				row.Cells[4].SetString(now.Name)
			case 2:

				var gradeMatch = false
				for _, g := range []string{"一年", "二年", "三年", "四年", "五年", "六年", "七年", "八年", "九年"} {
					if strings.Contains(now.Name, g) {
						gradeMatch = true
						break
					}
				}

				if !gradeMatch {
					slf.fullGrade = true
				}else {
					slice := strings.SplitN(now.Name, "年", 2)
					grade := slf.Stage + strings.ReplaceAll(slice[0], slf.Stage, "") + "年级"
					row.Cells[1].SetString(grade)
				}

				if !strings.Contains(now.Name, "册") {
					slf.fullBook = true
				}else {
					slice := strings.SplitN(now.Name, "年", 2)
					var book string

					var index = 1
					if !gradeMatch {
						index = 0
					}

					if strings.Contains(slice[index], "上") {
						book = "上册"
						row.Cells[3].SetString(book)
					} else if strings.Contains(slice[index], "下") {
						book = "下册"
						row.Cells[3].SetString(book)
					}else {
						slf.fullBook = true
					}
				}

			default:
				row.Cells[now.Level + 2].SetString(now.Name)
			}
			now = now.Parent
		}
		row.Cells[0].SetString(slf.Stage)
		row.Cells[2].SetString(slf.Subject)

		if !slf.fullGrade && slf.fullBook {
			slf.FullBook(row)
		}
		if slf.fullGrade && !slf.fullBook {
			slf.FullGrade(row)
		}
		if slf.fullGrade && slf.fullBook {
			slf.FullGradeAndBook(row)
		}
	}
}