package service

import (
	"fmt"
	"frame/orm"
	_ "github.com/go-sql-driver/mysql"
	"math/rand"
	"net/url"
)

//type User struct {
//	Id       int64  `gorm:"id,auto_increment"`
//	UserName string `gorm:"user_name"`
//	Password string `gorm:"password"`
//	Age      int    `gorm:"age"`
//}

// User 结构体（跟数据库中的结构一样）
type User struct {
	Id       int64
	UserName string
	Password string
	Age      int
}

// SaveUser 保存用户
func SaveUser() {
	dataSourceName := fmt.Sprintf("root:root@tcp(localhost:3306)/framego?charset=utf8&loc=%s&parseTime=true", url.QueryEscape("Asia/Shanghai"))
	db := orm.Open("mysql", dataSourceName)
	//db.Prefix = "framego_"
	user := &User{
		UserName: "yyds",
		Password: "123456",
		Age:      18,
	}
	id, _, err := db.New(&User{}).Insert(user)
	if err != nil {
		panic(err)
	}
	fmt.Println(id)

	db.Close()
}

func SaveUserBatch() {
	dataSourceName := fmt.Sprintf("root:root@tcp(localhost:3306)/framego?charset=utf8&loc=%s&parseTime=true", url.QueryEscape("Asia/Shanghai"))
	db := orm.Open("mysql", dataSourceName)
	//db.Prefix = "framego"
	user := &User{
		UserName: "yys1",
		Password: "12345612",
		Age:      36,
	}
	user1 := &User{
		UserName: "yys2",
		Password: "123456111",
		Age:      48,
	}
	var users []any
	users = append(users, user, user1)
	id, _, err := db.New(&User{}).InsertBatch(users)
	if err != nil {
		panic(err)
	}
	fmt.Println(id)

	db.Close()
}

func UpdateUser() {
	dataSourceName := fmt.Sprintf("root:root@tcp(localhost:3306)/framego?charset=utf8&loc=%s&parseTime=true", url.QueryEscape("Asia/Shanghai"))
	db := orm.Open("mysql", dataSourceName)
	//db.Prefix = "mframego_"
	//id, _, err := db.New().Where("id", 1006).Where("age", 54).Update(user)

	randInt := rand.Int()

	//单个插入
	user := &User{
		UserName: fmt.Sprintf("yyds%d", randInt),
		Password: "123456",
		Age:      30,
	}
	insertId, _, err := db.New(&User{}).Insert(user)
	if err != nil {
		panic(err)
	}
	fmt.Println(insertId)

	//批量插入
	var users []any
	//users = append(users, user)
	//id, _, err = db.New(&User{}).InsertBatch(users)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(id)
	//更新
	id, _, err := db.
		New(&User{}).
		Where("id", insertId).
		UpdateParam("age", 100).
		Update()
	//查询单行数据
	err = db.New(&User{}).
		Where("id", id).
		Or().
		Where("age", 30).
		SelectOne(user, "user_name")
	//查询多行数据
	users, err = db.New(&User{}).Select(&User{})
	if err != nil {
		panic(err)
	}
	for _, v := range users {
		u := v.(*User)
		fmt.Println(u)
	}

	if err != nil {
		panic(err)
	}
	fmt.Println(id)

	db.Close()
}

func SelectOne() {
	dataSourceName := fmt.Sprintf("root:root@tcp(localhost:3306)/framego?charset=utf8&loc=%s&parseTime=true", url.QueryEscape("Asia/Shanghai"))
	db := orm.Open("mysql", dataSourceName)
	//db.Prefix = "framego_"
	user := &User{}
	err := db.New(user).
		Where("id", 1).
		Or().
		Where("age", 36).
		SelectOne(user, "user_name")
	if err != nil {
		panic(err)
	}
	fmt.Println(user)

	db.Close()
}

func Select() {
	dataSourceName := fmt.Sprintf("root:root@tcp(localhost:3306)/framego?charset=utf8&loc=%s&parseTime=true", url.QueryEscape("Asia/Shanghai"))
	db := orm.Open("mysql", dataSourceName)
	//db.Prefix = "framego_"
	user := &User{}
	users, err := db.New(user).Order("id", "asc", "age", "desc").Select(user)
	if err != nil {
		panic(err)
	}
	for _, v := range users {
		u := v.(*User)
		fmt.Println(u)
	}
	db.Close()
}

func Count() {
	dataSourceName := fmt.Sprintf("root:root@tcp(localhost:3306)/framego?charset=utf8&loc=%s&parseTime=true", url.QueryEscape("Asia/Shanghai"))
	db := orm.Open("mysql", dataSourceName)
	//db.Prefix = "framego_"
	user := &User{}
	count, err := db.New(user).Count()
	if err != nil {
		panic(err)
	}
	fmt.Println(count)
	db.Close()
}
