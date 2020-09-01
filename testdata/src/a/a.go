package a

import (
	"encoding/json"
	"fmt"
)

type Student struct {
	Id    int
	Name  string
	Score int
}

func (s Student) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Id   int
		Name string
	}{
		s.Id,
		s.Name,
	})
}

type Teacher struct {
	Id     int
	Name   string
	Salary int
}

func (t *Teacher) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Id   int
		Name string
	}{
		t.Id,
		t.Name,
	})
}

func f() {
	var gopher int
	print(gopher)

	s := Student{1, "hoge", 100}
	t := Teacher{2, "fuga", 200}

	sStr, _ := json.Marshal(s)
	tStr, _ := json.Marshal(t) // want "NG"
	fmt.Println(t)

	fmt.Println(sStr)
	fmt.Println(tStr)
}
