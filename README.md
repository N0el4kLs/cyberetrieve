<h2 align="center">Cyber-Retrieve</h2>

## 介绍

Cyber-Retrieve 是使用Go实现的网络空间搜索引擎的第三方包, 用于从中检索信息.

1. 支持语法自动转换(部分)
2. 支持多种搜索引擎，目前已经集成

| 是否集成 |  名称   |                                                                                    地址                                                                                    |
|:----:|:-----:|:------------------------------------------------------------------------------------------------------------------------------------------------------------------------:|
|  ✅   | Quake |                    [https://quake.360.net](https://quake.360.net/quake/#/help?id=5e77423bcb9954d2f8a01656&title=%E4%BD%BF%E7%94%A8%E8%AF%B4%E6%98%8E)                    |
|  ✅   | Fofa  |                                                                [https://fofa.info](https://fofa.info/api)                                                                |

## 使用

安装
```shell
go get -u github.com/N0el4kLs/cyberetrieve
```

使用案例:
```go
import (
    "fmt"

    "github.com/N0el4kLs/cyberetrieve"
    "github.com/N0el4kLs/cyberetrieve/sources"
)

func main() {
	query := sources.Query{
		Query:         `title:"login"`,
		NumberOfQuery: 5,
	}
	session := sources.Session{
		QuakeToken: "xxx-xxx-xxx",
		FofaKey:    "xxx-xxx-xxx",
	}

	engine := cyberetrieve.NewCyberRetrieveEngine(query, session,
		cyberetrieve.WithAutoGrammar(),
		cyberetrieve.WithFofaSearch(),
	)
	if rst, err := engine.RetrieveResult(); err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println(rst)
	}
}
```
更多使用案例可以前往[example](./example)查看