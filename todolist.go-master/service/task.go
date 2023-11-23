package service

import (
	"fmt"
	"net/http"
	"strconv"

	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	database "todolist.go/db"
)

// TaskList renders list of tasks in DB
func TaskList(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	var tags []database.Tag

	err = db.Select(&tags, "SELECT tag_name FROM tags")
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// Get query parameter
	kw := ctx.Query("kw")
	done_check := ctx.Query("done_check")
	tag_se := ctx.Query("tag_se")

	// Get tasks in DB
	var tasks []database.Task
	if tag_se != "" && tag_se != "指定しない" {
		query := "SELECT id, title, created_at, due_to, priority, is_done FROM tasks INNER JOIN ownership ON task_id = id WHERE user_id = ? AND tag = ?"
		if done_check == "未実行" {
			switch {
			case kw != "":
				err = db.Select(&tasks, query+" AND title LIKE ? AND is_done = false", userID, tag_se, "%"+kw+"%")
			default:
				err = db.Select(&tasks, query+" AND is_done = false", userID, tag_se)
			}
			if err != nil {
				Error(http.StatusInternalServerError, err.Error())(ctx)
				return
			}
		} else if done_check == "実行済" {
			switch {
			case kw != "":
				err = db.Select(&tasks, query+" AND title LIKE ? AND is_done = true", userID, tag_se, "%"+kw+"%")
			default:
				err = db.Select(&tasks, query+" AND is_done = true", userID, tag_se)
			}
			if err != nil {
				Error(http.StatusInternalServerError, err.Error())(ctx)
				return
			}
		} else {
			switch {
			case kw != "":
				err = db.Select(&tasks, query+" AND title LIKE ?", userID, tag_se, "%"+kw+"%")
			default:
				err = db.Select(&tasks, query, userID, tag_se)
			}
			if err != nil {
				Error(http.StatusInternalServerError, err.Error())(ctx)
				return
			}
		}
	} else {
		query := "SELECT id, title, created_at, due_to, priority, is_done FROM tasks INNER JOIN ownership ON task_id = id WHERE user_id = ?"
		if done_check == "未実行" {
			switch {
			case kw != "":
				err = db.Select(&tasks, query+" AND title LIKE ? AND is_done = false", userID, "%"+kw+"%")
			default:
				err = db.Select(&tasks, query+" AND is_done = false", userID)
			}
			if err != nil {
				Error(http.StatusInternalServerError, err.Error())(ctx)
				return
			}
		} else if done_check == "実行済" {
			switch {
			case kw != "":
				err = db.Select(&tasks, query+" AND title LIKE ? AND is_done = true", userID, "%"+kw+"%")
			default:
				err = db.Select(&tasks, query+" AND is_done = true", userID)
			}
			if err != nil {
				Error(http.StatusInternalServerError, err.Error())(ctx)
				return
			}
		} else {
			switch {
			case kw != "":
				err = db.Select(&tasks, query+" AND title LIKE ?", userID, "%"+kw+"%")
			default:
				err = db.Select(&tasks, query, userID)
			}
			if err != nil {
				Error(http.StatusInternalServerError, err.Error())(ctx)
				return
			}
		}
	}
	t := time.Now()
	for i := 0; i < len(tasks); i++ {
		tasks[i].DueTo_Str = tasks[i].DueTo.Format("2006/01/02")
		if tasks[i].DueTo_Str == "0001/01/01" {
			tasks[i].DueTo_Str = "未設定"
		}
		tasks[i].CreatedAt_Str = tasks[i].CreatedAt.Format("2006/1/2 15:04:05")
		tasks[i].RestDay = int((tasks[i].DueTo.Sub(t).Hours() + 24) / 24)
	}

	// Render tasks
	ctx.HTML(http.StatusOK, "task_list.html", gin.H{"Title": "Task list", "Tasks": tasks, "Kw": kw, "Done_check": done_check, "Tag_se": tag_se, "Tags": tags})
}

// ShowTask renders a task with given ID
func ShowTask(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// parse ID given as a parameter
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	/*var username string
	err = db.Get(&username, "SELECT name FROM users WHERE id=?", userID) // Use DB#Get for one entry
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}*/

	var user_id []uint64
	err = db.Select(&user_id, "SELECT user_id FROM ownership WHERE task_id=?", id) // Use DB#Get for one entry
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	// Get a task with given ID
	var task database.Task
	err = db.Get(&task, "SELECT * FROM tasks WHERE id=?", id) // Use DB#Get for one entry
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	var createname string
	err = db.Get(&createname, "SELECT name FROM users WHERE id=?", task.CreateUser) // Use DB#Get for one entry
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	task.DueTo_Str = task.DueTo.Format("2006-01-02")
	if task.DueTo_Str == "0001-01-01" {
		task.DueTo_Str = "未設定"
	}
	count := 0
	for _, user_ids := range user_id {
		if userID.(uint64) == user_ids {
			count++
		}
	}
	if count == 0 {
		ctx.Redirect(http.StatusFound, "/login")
		return
	}

	// Render task
	//ctx.String(http.StatusOK, task.Title)  // Modify it!!
	ctx.HTML(http.StatusOK, "task.html", gin.H{"Task": task, "CreateUser": createname, "MyUserID": userID})
}

func NewTaskForm(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")
	db, err := database.GetConnection()
	var tags []database.Tag
	var users []database.User
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

	err = db.Select(&tags, "SELECT tag_name FROM tags")
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	err = db.Select(&users, "SELECT name FROM users")
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	ctx.HTML(http.StatusOK, "form_new_task.html", gin.H{"Title": "Task registration", "Tags": tags, "Users": users, "UserName": username})
}

func RegisterTask(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")
	// Get task title
	title, exist := ctx.GetPostForm("title")
	if !exist {
		Error(http.StatusBadRequest, "No title is given")(ctx)
		return
	}
	description, _ := ctx.GetPostForm("description")
	due_to, _ := ctx.GetPostForm("due_to")
	priority, _ := ctx.GetPostForm("priority")
	tag, _ := ctx.GetPostForm("tag")
	new_tag, new_tag_ex := ctx.GetPostForm("new_tag")
	shareuser, _ := ctx.GetPostFormArray("share")
	ex_date := due_to == ""
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	tx := db.MustBegin()
	if new_tag_ex && new_tag != "" {
		var duplicate int
		err = db.Get(&duplicate, "SELECT COUNT(*) FROM tags WHERE tag_name=?", new_tag)
		if err != nil {
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
		}
		if duplicate > 0 {
			var tags []database.Tag
			var users []database.User

			err = db.Select(&tags, "SELECT tag_name FROM tags")
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
			err = db.Select(&users, "SELECT name FROM users")
			if err != nil {
				Error(http.StatusInternalServerError, err.Error())(ctx)
				return
			}
			if !ex_date {
				ctx.HTML(http.StatusBadRequest, "form_new_task.html", gin.H{"Title": "Task registration", "Error": "tagname is already taken", "Tags": tags, "TaskTitle": title, "Priority": priority, "Explain": description, "DueTo": due_to, "Users": users, "UserName": username, "ShareUser": shareuser})
			} else {
				ctx.HTML(http.StatusBadRequest, "form_new_task.html", gin.H{"Title": "Task registration", "Error": "tagname is already taken", "Tags": tags, "TaskTitle": title, "Priority": priority, "Explain": description, "Users": users, "UserName": username, "ShareUser": shareuser})
			}
			return
		}
		_, err := tx.Exec("INSERT INTO tags (tag_name) VALUES (?)", new_tag)
		if err != nil {
			tx.Rollback()
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
		}
		tag = new_tag
	}
	var taskID int64
	if !ex_date {
		result, err := tx.Exec("INSERT INTO tasks (title , explanation , due_to , priority , tag, create_user) VALUES (? ,? ,? ,? ,? ,?)", title, description, due_to, priority, tag, userID)
		if err != nil {
			tx.Rollback()
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
		}
		taskID, err = result.LastInsertId()
		if err != nil {
			tx.Rollback()
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
		}
	} else {
		result, err := tx.Exec("INSERT INTO tasks (title , explanation , due_to , priority , tag, create_user) VALUES (? ,? ,? ,? ,? ,?)", title, description, "0001-01-01", priority, tag, userID)
		if err != nil {
			tx.Rollback()
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
		}
		taskID, err = result.LastInsertId()
		if err != nil {
			tx.Rollback()
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
		}
	}
	_, err = tx.Exec("INSERT INTO ownership (user_id, task_id) VALUES (?, ?)", userID, taskID)
	if err != nil {
		tx.Rollback()
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	//names := ""
	for _, username := range shareuser {
		if username != "未設定" {
			var userid int64
			err = tx.Get(&userid, "SELECT id FROM users WHERE name=?", username) // Use DB#Get for one entry
			if err != nil {
				tx.Rollback()
				Error(http.StatusBadRequest, err.Error())(ctx)
				return
			}
			_, err = tx.Exec("INSERT INTO ownership (user_id, task_id) VALUES (?, ?)", userid, taskID)
			if err != nil {
				tx.Rollback()
				Error(http.StatusInternalServerError, err.Error())(ctx)
				return
			}
		}
	}
	tx.Commit()
	//ctx.Redirect(http.StatusFound, fmt.Sprintf("/task/%d", taskID))
	ctx.Redirect(http.StatusFound, fmt.Sprintf("/task/%d", taskID))
}

func EditTaskForm(ctx *gin.Context) {
	// ID の取得
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	userID := sessions.Default(ctx).Get("user")
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	var tags []database.Tag
	var users []database.User

	err = db.Select(&tags, "SELECT tag_name FROM tags")
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	// Get target task
	var task database.Task
	err = db.Get(&task, "SELECT * FROM tasks WHERE id=?", id)
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	task.DueTo_Str = task.DueTo.Format("2006-01-02")
	err = db.Select(&users, "SELECT name FROM users")
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
	var shareuser_id []int64
	err = db.Select(&shareuser_id, "SELECT user_id FROM ownership WHERE task_id = ?", id)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	var shareusers []string
	for _, suser_id := range shareuser_id {
		var s_name string
		err = db.Get(&s_name, "SELECT name FROM users WHERE id=?", suser_id) // Use DB#Get for one entry
		if err != nil {
			Error(http.StatusBadRequest, err.Error())(ctx)
			return
		}
		shareusers = append(shareusers, s_name)
	}
	// Render edit form
	ctx.HTML(http.StatusOK, "form_edit_task.html",
		gin.H{"Title": fmt.Sprintf("Edit task %d", task.ID), "Task": task, "Tags": tags, "Users": users, "MyUserID": userID, "UserName": username, "ShareUser": shareusers})
}

func UpdateTask(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	// Get task title
	title, exist := ctx.GetPostForm("title")
	if !exist {
		Error(http.StatusBadRequest, "No title is given")(ctx)
		return
	}
	is_done, exist := ctx.GetPostForm("is_done")
	if !exist {
		Error(http.StatusBadRequest, "No is_done is given")(ctx)
		return
	}
	description, _ := ctx.GetPostForm("description")
	due_to, _ := ctx.GetPostForm("due_to")
	tag, _ := ctx.GetPostForm("tag")
	priority, _ := ctx.GetPostForm("priority")
	shareuser, _ := ctx.GetPostFormArray("share")
	new_tag, new_tag_ex := ctx.GetPostForm("new_tag")
	ex_date := due_to == ""
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	// Create new data with given title on DB
	//result, err := db.Exec("UPDATE tasks SET (title) = (?), (is_done) = (?) WHERE id = ?", title, is_done, id)
	done_check, err := strconv.ParseBool(is_done)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	if new_tag_ex && new_tag != "" {
		var duplicate int
		err = db.Get(&duplicate, "SELECT COUNT(*) FROM tags WHERE tag_name=?", new_tag)
		if err != nil {
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
		}
		if duplicate > 0 {
			var task database.Task
			err = db.Get(&task, "SELECT * FROM tasks WHERE id=?", id)
			if err != nil {
				Error(http.StatusBadRequest, err.Error())(ctx)
				return
			}
			var tags []database.Tag
			var users []database.User

			err = db.Select(&tags, "SELECT tag_name FROM tags")
			if err != nil {
				Error(http.StatusInternalServerError, err.Error())(ctx)
				return
			}
			task.DueTo_Str = task.DueTo.Format("2006-01-02")
			var username string
			err = db.Get(&username, "SELECT name FROM users WHERE id=?", userID) // Use DB#Get for one entry
			if err != nil {
				Error(http.StatusBadRequest, err.Error())(ctx)
				return
			}
			var shareuser_id []int64
			err = db.Select(&shareuser_id, "SELECT user_id FROM ownership WHERE task_id = ?", id)
			if err != nil {
				Error(http.StatusInternalServerError, err.Error())(ctx)
				return
			}
			var shareusers []string
			for _, suser_id := range shareuser_id {
				var s_name string
				err = db.Get(&s_name, "SELECT name FROM users WHERE id=?", suser_id) // Use DB#Get for one entry
				if err != nil {
					Error(http.StatusBadRequest, err.Error())(ctx)
					return
				}
				shareusers = append(shareusers, s_name)
			}
			ctx.HTML(http.StatusBadRequest, "form_edit_task.html", gin.H{"Title": fmt.Sprintf("Edit task %d", task.ID), "Error": "tagname is already taken", "Task": task, "Tags": tags, "Users": users, "MyUserID": userID, "UserName": username, "ShareUser": shareusers})
			return
		}
		_, err := db.Exec("INSERT INTO tags (tag_name) VALUES (?)", new_tag)
		if err != nil {
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
		}
		tag = new_tag
	}
	if !ex_date {
		_, erro := db.Exec("UPDATE tasks SET title = ?, is_done = ?, explanation = ?, due_to = ?, priority = ?, tag = ?  WHERE id = ?", title, done_check, description, due_to, priority, tag, id)
		if erro != nil {
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
		}
	} else {
		_, erro := db.Exec("UPDATE tasks SET title = ?, is_done = ?, explanation = ?, due_to = ?, priority = ?, tag = ?  WHERE id = ?", title, done_check, description, "0001-01-01", priority, tag, id)
		if erro != nil {
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
		}
	}

	_, err = db.Exec("DELETE FROM ownership WHERE task_id=?", id)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	_, err = db.Exec("INSERT INTO ownership (user_id, task_id) VALUES (?, ?)", userID, id)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	for _, username := range shareuser {
		if username != "未設定" {
			var userid int64
			err = db.Get(&userid, "SELECT id FROM users WHERE name=?", username) // Use DB#Get for one entry
			if err != nil {
				Error(http.StatusBadRequest, err.Error())(ctx)
				return
			}
			_, err = db.Exec("INSERT INTO ownership (user_id, task_id) VALUES (?, ?)", userid, id)
			if err != nil {
				Error(http.StatusInternalServerError, err.Error())(ctx)
				return
			}
		}
	}
	// Render status
	path := fmt.Sprintf("/task/%d", id) ///task/<id> へ戻る
	ctx.Redirect(http.StatusFound, path)
}

func DeleteTask(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")
	// ID の取得
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	// Get DB connection
	//var task database.Task
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	var task database.Task
	err = db.Get(&task, "SELECT * FROM tasks WHERE id=?", id)
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	if task.CreateUser == userID {
		// Delete the task from DB
		_, err = db.Exec("DELETE FROM tasks WHERE id=?", id)
		if err != nil {
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
		}
	} else {
		_, err = db.Exec("DELETE FROM ownership WHERE task_id=? AND user_id=?", id, userID)
		if err != nil {
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
		}
	}
	// Redirect to /list
	ctx.Redirect(http.StatusFound, "/list")
}
