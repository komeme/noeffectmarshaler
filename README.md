# noeffectmarshaler

## What's this?

構造体json.Marshalerを実装したつもりになっていが、json.Marshalに対してstructを値として渡してしまうことにより、json.Marshalerを実装したとみなされないパターンの検出
（json.Marshalからすればjson.Marshalerを実装するかどうかは任意なのでエラーを発生させない）

（特定の構造体が特定の関数の引数に入れられるのを検出する。という拡張もアリかな）

```go
package main

import (
	"encoding/json"
	"fmt"
)

type Person struct {
	Name  string
	Age   int
	Score int
}

func (p *Person) MarshalJSON() ([]byte, error) {
	return json.Marshal("censored")
}

func main() {
	p := Person{
		Name:  "Hoge",
		Age:   25,
		Score: -100,
	}

	data, err := json.Marshal(p)
	if err != nil{
		panic(err)
	}

	fmt.Println(string(data)) // "{"Name":"Hoge","Age":25,"Score":-100}" printed 
}

```

## Install
```
[WIP]
```

## Usage
```
[WIP]
```

## おまけ
- そもそも構造体をポインタではなく値として変数に入れること自体をやめるべきであるとうい指摘があった。