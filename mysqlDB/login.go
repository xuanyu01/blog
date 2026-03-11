package mysqlDB

import (
	"blog/redisDB"
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

// 未加密密码,暂时以明文存储
func LoginUser(db *sql.DB, username string, password string) string {

	loginQuery := "select password from user where username=? "

	var storedPassword string
	err := db.QueryRow(loginQuery, username).Scan(&storedPassword)
	if err != nil { //用户不存在或查询出错
		log.Println("Error querying user:", err)
		return "Username or Password is Invalid"
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))
	if err != nil {
		//密码不匹配
		log.Println("password is wrong", err)
		return "Username or Password is Invalid"
	}
	return "Success Login"
}

// 检测是否登录
func IndexHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		blogs, _ := GetBlogs(db)

		isLogin := false
		userImage := ""

		sessionID, err := c.Cookie("sessionID")
		if err == nil {
			log.Println(err)

			userID, err := redisDB.GetSession(sessionID)
			if err == nil {

				isLogin = true

				// 从数据库读取用户头像
				userImage, _ = GetUserImage(db, userID)
			}
		}

		c.HTML(200, "index.html", gin.H{
			"title":     "Xuan",
			"blogs":     blogs,
			"isLogin":   isLogin,
			"UserImage": userImage,
		})
	}
}

// 用户登出
func UserLogout(c *gin.Engine) bool {
	logout := true
	c.GET("/logout", func(c *gin.Context) {
		sessionID, err := c.Cookie("session_id")
		if err == nil {
			logout = false
			log.Fatal(err)
		}
		redisDB.DeleteSession(sessionID)
		c.SetCookie("session_id", "", -1, "/", "", false, true)
		c.Redirect(302, "/")
	})
	if logout == true {
		return true
	}
	if logout == false {
		return false
	}
	return false
}
