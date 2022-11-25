package main

import (
	"database/sql"
	"github.com/gin-gonic/gin"

	//"fmt"
	//"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

//登录功能：
/*
1.登录时，首先检查是否有cookie，有cookie且正确就维持状态，没有cookie就提示输入密码，有cookie但错误就清楚cookie
2.注册接口：主要检验账号名是否存在，密码是否符合要求，都满足就入表，同时需要在注册时提供密保问题
3.登录接口：在没有cookie的情况时，客户端输入账号，密码，服务端校验账号密码是否已经存在，如果不存在就添加进入数据表，如果存在就直接登录
4.忘记密码接口：通过查表获取密保，找回密码
5.由以上功能，定义数据表的结构为id，name，password，密保问题1，消息栏（有他人发的消息就往这个栏append，通过id标识特定的人）
*/

/*
项目结构：
/根是用户界面，需要对cookie校验
0.主函数：接口注册
1.cookie校验函数
2.注册函数集
3.登录函数集
4.忘记密码函数集
5.留言函数集
5.注册接口
6.登录接口
7.忘记密码接口
8.留言接口
*/

/*
其他函数
1.表的初始化
2.线程池的使用

*/

// cookie判断，一旦检测到cookie缺失或者错误（账号和密码对不上），就返回cookie不正确，要求重新登录
// 如果cookie都存在且正确，就继续下面的逻辑
/*
建表：
CREATE TABLE `user` (
`id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
`username` varCHAR(16) NOT NULL UNIQUE,
`password` varchar(20) NOT NULL,
`question` varchar(20) NOT NULL,
`question_content` varchar(50) NOT NULL,
`message`  varchar(1000) NOT NULL default "no message"
);




*/

type user struct {
	id               int
	name             string
	password         string
	question         string
	question_content string
	message          string
}

// 检查cookie，一旦有误一律返回错误，让用户重新登录
func cookie_check(db *sql.DB, username string, err1 error) (map[string]interface{}, error) {
	//第一种情况，没有cookie时
	if err1 != nil {
		//如果cookie为空，返回错误信息
		return map[string]interface{}{"ok": false, "data": "no cookie"}, err1
	} else { //第二种情况：提供了cookie
		datalist, err := selectFromUserName(db, username)
		//当查不到此用户，说明cookie有误
		if err != nil {
			return map[string]interface{}{"ok": false, "data": "no such username"}, err
		}
		//提取列表的第一条，这就是查询的信息
		data := datalist[0]
		//校验成功
		if data.name == username {
			return map[string]interface{}{"ok": true, "data": "check_success"}, nil
		}
	} //默认返回
	return nil, err1
}

// 进入数据库查用户的账号密码是否正确，在这里直接拿做登录接口了
func check_user_pass(db *sql.DB, username string, password string) bool {
	datalist, err := selectFromUserName(db, username)
	if err != nil {
		return false
	}
	data := datalist[0]
	if password == data.password {
		return true
	}
	return false
}

// 注册，提供用户基础信息，完成注册
// 先检验是否存在，不存在就创建
func register(db *sql.DB, username string, password string, question string, question_content string) (bool, string) {
	//当不存在时，创建用户
	ok := insert_data(db, username, password, question, question_content)
	if !ok {
		return ok, "插入失败，可能已经存在该用户"
	}
	return true, "插入成功"
}

func getDb() (*sql.DB, error) {
	var dns = "root:chrnbfj666@tcp(127.0.0.1:3306)/test"
	db, err := sql.Open("mysql", dns)
	if err != nil {
		return nil, err
	}
	return db, nil
}

//查询数据，返回结构体数组（多条数据），和错误处理

func selectFromUserName(db *sql.DB, username string) ([]user, error) {
	var data user
	var list []user
	rows, err := db.Query("select * from user where username=?", username)
	if err != nil {
		return []user{}, err
	}
	defer rows.Close()
	for rows.Next() {
		// row.scan 必须按照先后顺序 &获取数据
		//var temp interface{}
		//由于表的结构是固定的，这里可以写死格式
		err := rows.Scan(&data.id, &data.name, &data.password, &data.question, &data.question_content, &data.message)
		if err != nil {
			log.Println(err)
			return []user{}, err
		}
		list = append(list, data)
	}
	return list, nil
}

// 插入一条数据，返回是否插入成功
func insert_data(db *sql.DB, username string, password string, question string, question_content string) bool {
	_, err := db.Exec("insert into user (username ,password ,question, question_content) value (?,?,?,?)", username, password, question, question_content)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}
func root(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		cuname, err1 := c.Cookie("username")
		//只有当cookie username和password同时存在时
		resp, err := cookie_check(db, cuname, err1)
		//当cookie有误时，返回json
		if err != nil {
			c.JSON(200, map[string]interface{}{"ok": false, "data": resp})
		}
		c.JSON(200, map[string]interface{}{"ok": true, "data": "你好，用户" + cuname})
	}
}
func regis(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.DefaultQuery("username", "null")
		password := c.DefaultQuery("password", "null")
		question := c.DefaultQuery("question", "null")
		question_content := c.DefaultQuery("question_content", "null")
		//当用户名和密码有任何一部分没有被提供时，返回信息
		if username == "null" || password == "null" || question == "null" || question_content == "null" {
			c.JSON(200, map[string]interface{}{"ok": false, "data": "lack of content"})
		}
		//当全部内容都非空时，注册用户
		ok, _ := register(db, username, password, question, question_content)
		if ok {
			//设置cookie在根
			c.SetCookie("username", username, 3600, "/", "localhost", false, true)
			c.JSON(200, map[string]interface{}{"ok": true, "data": "用户创建成功，接下来返回主页"})
			c.Redirect(303, "/")
		} else {
			c.JSON(200, map[string]interface{}{"ok": false, "data": "用户创建失败，也许用户已经存在了罢"})
		}
	}
}
func login(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.DefaultQuery("username", "null")
		password := c.DefaultQuery("password", "null")
		if username == "null" || password == "null" {
			c.JSON(200, map[string]interface{}{"ok": false, "data": "no username or password"})
		}

		ok := check_user_pass(db, username, password)
		if ok {
			c.SetCookie("username", username, 3600, "/", "localhost", false, true)
			c.JSON(200, map[string]interface{}{"ok": true, "data": "用户登录成功，接下来返回主页"})
			c.Redirect(303, "/")
		} else {
			c.JSON(200, map[string]interface{}{"ok": false, "data": "登录失败，用户不存在或密码错误"})
		}
	}
}
func forpost(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.DefaultQuery("username", "null")
		question_content := c.DefaultQuery("question_content", "null")
		if username == "null" || question_content == "null" {
			c.JSON(200, map[string]interface{}{"ok": false, "data": "lack of content"})
		}
		datalist, err := selectFromUserName(db, username)
		if err != nil {
			c.JSON(200, map[string]interface{}{"ok": false, "data": "get data wrong"})
		}
		data := datalist[0]
		if data.question_content == question_content {
			c.JSON(200, map[string]interface{}{"ok": true, "data": "你的密码是" + data.password + "\n下次别再忘了哦"})
		} else {
			c.JSON(200, map[string]interface{}{"ok": false, "data": "密保问题错误啦"})
		}

	}
}
func forget(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.DefaultQuery("username", "null")
		if username == "null" {
			c.JSON(200, map[string]interface{}{"ok": false, "data": "lack of content"})
		}
		datalist, err := selectFromUserName(db, username)
		if err != nil {
			c.JSON(200, map[string]interface{}{"ok": false, "data": "get data wrong"})
		} else { //存在则返回用户密保
			data := datalist[0]
			c.JSON(200, map[string]interface{}{"ok": true, "data": data.question})
		}
	}
}

func main() {
	db, err := getDb()
	if err != nil {
		return
	}

	router := gin.Default()
	//用户主页，欢迎用户
	router.GET("/", root(db), func(c *gin.Context) {
	})
	router.POST("/register", regis(db), func(c *gin.Context) {
	})
	router.POST("/login", login(db), func(c *gin.Context) {
	})
	router.POST("/forget", forpost(db), func(c *gin.Context) {

	})
	//相同地址，获取用户密保问题，提供给前端调用，data中保存密保问题
	router.GET("/forget", forget(db), func(c *gin.Context) {
	})
	//router.POST("/leave_message", func(c *gin.Context) {
	//
	//})
	router.Run(":5000")
}
