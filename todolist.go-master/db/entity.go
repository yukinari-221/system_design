package db

// schema.go provides data models in DB
import (
	"time"
)

// Task corresponds to a row in `tasks` table
type Task struct {
	ID            uint64    `db:"id"`
	Title         string    `db:"title"`
	CreatedAt     time.Time `db:"created_at"`
	IsDone        bool      `db:"is_done"`
	Explanation   string    `db:"explanation"`
	DueTo         time.Time `db:"due_to"`
	Priority      string    `db:"priority"`
	Tag           string    `db:"tag"`
	CreateUser    uint64    `db:"create_user"`
	DueTo_Str     string
	CreatedAt_Str string
	RestDay       int
}

type User struct {
	ID       uint64 `db:"id"`
	Name     string `db:"name"`
	Password []byte `db:"password"`
}

type Ownership struct {
	User_id uint64 `db:"user_id"`
	Task_id uint64 `db:"task_id"`
}

type Tag struct {
	Tag_name string `db:"tag_name"`
}
