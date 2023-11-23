package service

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	database "todolist.go/db"
)

func NewUserForm(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "new_user_form.html", gin.H{"Title": "Register user"})
}

func hash(pw string) []byte {
	const salt = "todolist.go#"
	h := sha256.New()
	h.Write([]byte(salt))
	h.Write([]byte(pw))
	return h.Sum(nil)
}

func RegisterUser(ctx *gin.Context) {
	// フォームデータの受け取り
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")
	password_check := ctx.PostForm("password_check")
	switch {
	case username == "":
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Usernane is not provided", "Username": username})
	case password == "":
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password is not provided", "Password": password})
	case password_check == "":
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password is not provided", "Password_check": password_check})
	}

	if password != password_check {
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Passwords do not match", "Username": username, "Password": password, "Password_check": password_check})
		return
	}

	//パスワードの複雑さを確認
	num_bool := NumCheck(password)
	upper_bool := UpperCheck(password)
	lower_bool := LowerCheck(password)

	if !num_bool {
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password does not contain numbers", "Username": username, "Password": password, "Password_check": password_check})
		return
	}
	if !upper_bool {
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password does not contain uppercase letters", "Username": username, "Password": password, "Password_check": password_check})
		return
	}
	if !lower_bool {
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password does not contain lowercase letters", "Username": username, "Password": password, "Password_check": password_check})
		return
	}

	// DB 接続
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// 重複チェック
	var duplicate int
	err = db.Get(&duplicate, "SELECT COUNT(*) FROM users WHERE name=?", username)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	if duplicate > 0 {
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Username is already taken", "Username": username, "Password": password})
		return
	}

	// DB への保存
	_, err = db.Exec("INSERT INTO users(name, password) VALUES (?, ?)", username, hash(password))
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	ctx.Redirect(http.StatusFound, "/login")
}

func NumCheck(str string) bool {
	for _, r := range str {
		if '0' <= r && r <= '9' {
			return true
		}
	}
	return false
}
func UpperCheck(str string) bool {
	for _, r := range str {
		if 'A' <= r && r <= 'Z' {
			return true
		}
	}
	return false
}

func LowerCheck(str string) bool {
	for _, r := range str {
		if 'a' <= r && r <= 'z' {
			return true
		}
	}
	return false
}

func LoginForm(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "login.html", gin.H{"Title": "Login"})
}

const userkey = "user"

func Login(ctx *gin.Context) {
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")

	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// ユーザの取得
	var user database.User
	err = db.Get(&user, "SELECT id, name, password FROM users WHERE name = ?", username)
	if err != nil {
		ctx.HTML(http.StatusBadRequest, "login.html", gin.H{"Title": "Login", "Username": username, "Error": "No such user"})
		return
	}

	// パスワードの照合
	if hex.EncodeToString(user.Password) != hex.EncodeToString(hash(password)) {
		ctx.HTML(http.StatusBadRequest, "login.html", gin.H{"Title": "Login", "Username": username, "Error": "Incorrect password"})
		return
	}

	// セッションの保存
	session := sessions.Default(ctx)
	session.Set(userkey, user.ID)
	session.Save()

	ctx.Redirect(http.StatusFound, "/list")
}

func LoginCheck(ctx *gin.Context) {
	if sessions.Default(ctx).Get(userkey) == nil {
		ctx.Redirect(http.StatusFound, "/login")
		ctx.Abort()
	} else {
		ctx.Next()
	}
}

func Logout(ctx *gin.Context) {
	session := sessions.Default(ctx)
	session.Clear()
	session.Options(sessions.Options{MaxAge: -1})
	session.Save()
	ctx.Redirect(http.StatusFound, "/")
}

func DeleteUser(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")
	password := ctx.PostForm("password")
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// パスワードの取得
	var pass []byte
	err = db.Get(&pass, "SELECT password FROM users WHERE id = ?", userID)
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	var username string
	err = db.Get(&username, "SELECT name FROM users WHERE id=?", userID) // Use DB#Get for one entry
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	// パスワードの照合
	if hex.EncodeToString(pass) != hex.EncodeToString(hash(password)) {
		ctx.HTML(http.StatusBadRequest, "delete_user.html", gin.H{"Title": "Delete user", "Error": "Incorrect password", "Username": username})
		return
	}

	var task_id uint64
	task_ids, err := db.Query("SELECT task_id FROM ownership WHERE user_id=?", userID) // Use DB#Get for one entry
	if err != nil {
		Error(http.StatusBadRequest, err.Error()+"userid :"+strconv.FormatUint(userID.(uint64), 10))(ctx)
		return
	}
	for task_ids.Next() {
		err := task_ids.Scan(&task_id)
		if err != nil {
			Error(http.StatusInternalServerError, err.Error()+"userid1 :"+strconv.FormatUint(userID.(uint64), 10))(ctx)
			return
		}
		_, err = db.Exec("DELETE FROM tasks WHERE id = ?", task_id)
		if err != nil {
			Error(http.StatusInternalServerError, err.Error()+"userid2 :"+strconv.FormatUint(userID.(uint64), 10))(ctx)
			return
		}
	}
	_, _ = db.Exec("DELETE FROM ownership WHERE user_id=?", userID)
	_, _ = db.Exec("DELETE FROM users WHERE id=?", userID)
	session := sessions.Default(ctx)
	session.Clear()
	session.Options(sessions.Options{MaxAge: -1})
	session.Save()
	ctx.Redirect(http.StatusFound, "/")
}

func DeleteForm(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	var username string
	err = db.Get(&username, "SELECT name FROM users WHERE id=?", userID) // Use DB#Get for one entry
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	ctx.HTML(http.StatusOK, "delete_user.html", gin.H{"Title": "Delete user", "Username": username})
}

func ChangeForm(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	var username string
	err = db.Get(&username, "SELECT name FROM users WHERE id=?", userID) // Use DB#Get for one entry
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	ctx.HTML(http.StatusOK, "change_user.html", gin.H{"Title": "Change user info", "Username": username})
}

func ChangeUser(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")
	// フォームデータの受け取り
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")
	password_check := ctx.PostForm("password_check")
	password_now := ctx.PostForm("password_now")
	switch {
	case username == "":
		ctx.HTML(http.StatusBadRequest, "change_user.html", gin.H{"Title": "Change user info", "Error": "Usernane is not provided", "Username": username})
	case password == "":
		ctx.HTML(http.StatusBadRequest, "change_user.html", gin.H{"Title": "Change user info", "Error": "Password is not provided"})
	case password_check == "":
		ctx.HTML(http.StatusBadRequest, "change_user.html", gin.H{"Title": "Change user info", "Error": "Password_Check is not provided"})
	case password_now == "":
		ctx.HTML(http.StatusBadRequest, "change_user.html", gin.H{"Title": "Change user info", "Error": "Current Password is not provided"})
	}

	if password != password_check {
		ctx.HTML(http.StatusBadRequest, "change_user.html", gin.H{"Title": "Change user info", "Error": "Passwords do not match", "Username": username})
		return
	}

	//パスワードの複雑さを確認
	num_bool := NumCheck(password)
	upper_bool := UpperCheck(password)
	lower_bool := LowerCheck(password)

	if !num_bool {
		ctx.HTML(http.StatusBadRequest, "change_user.html", gin.H{"Title": "Change user info", "Error": "Password does not contain numbers", "Username": username})
		return
	}
	if !upper_bool {
		ctx.HTML(http.StatusBadRequest, "change_user.html", gin.H{"Title": "Change user info", "Error": "Password does not contain uppercase letters", "Username": username})
		return
	}
	if !lower_bool {
		ctx.HTML(http.StatusBadRequest, "change_user.html", gin.H{"Title": "Change user info", "Error": "Password does not contain lowercase letters", "Username": username})
		return
	}

	// DB 接続
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// パスワードの取得
	var pass []byte
	err = db.Get(&pass, "SELECT password FROM users WHERE id = ?", userID)
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	var username_now string
	err = db.Get(&username_now, "SELECT name FROM users WHERE id=?", userID) // Use DB#Get for one entry
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	// パスワードの照合
	if hex.EncodeToString(pass) != hex.EncodeToString(hash(password_now)) {
		ctx.HTML(http.StatusBadRequest, "change_user.html", gin.H{"Title": "Change user info", "Error": "Incorrect password", "Username": username})
		return
	}

	// 重複チェック
	if username != username_now { //現在と同じ名前を使うのはOK
		var duplicate int
		err = db.Get(&duplicate, "SELECT COUNT(*) FROM users WHERE name=?", username)
		if err != nil {
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
		}
		if duplicate > 0 {
			ctx.HTML(http.StatusBadRequest, "change_user.html", gin.H{"Title": "Register user", "Error": "Username is already taken", "Username": username_now})
			return
		}
	}

	// DB への保存
	_, err = db.Exec("UPDATE users SET name = ?, password = ? WHERE id = ?", username, hash(password), userID)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	session := sessions.Default(ctx)
	session.Clear()
	session.Options(sessions.Options{MaxAge: -1})
	session.Save()
	ctx.Redirect(http.StatusFound, "/login")
}
