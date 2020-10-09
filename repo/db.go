package repo

import (
	"fmt"
	"time"

	"github.com/team4yf/fpm-go-plugin-cron/model"
	"github.com/team4yf/yf-fpm-server-go/pkg/db"
)

type dbJobRepo struct {
	dbclient db.Database
}

func (r *dbJobRepo) List() ([]*model.Job, error) {
	list := make([]*model.Job, 0)
	q := db.NewQuery()
	q.SetTable(model.Job{}.TableName())
	err := r.dbclient.Find(q, &list)
	return list, err
}

func (r *dbJobRepo) Tasks(code string, skip, limit int) ([]*model.Task, int, error) {
	list := make([]*model.Task, 0)
	q := db.NewQuery()
	q.SetTable(model.Task{}.TableName())
	q.SetCondition("code = ?", code)
	q.SetPager(&db.Pagination{
		Skip:  skip,
		Limit: limit,
	})
	q.AddSorter(db.Sorter{
		Sortby: "start_at",
		Asc:    "desc",
	})
	var total int64
	err := r.dbclient.FindAndCount(q, &list, &total)
	return list, (int)(total), err
}

func (r *dbJobRepo) CreateJob(j *model.Job) (err error) {
	q := db.NewQuery()
	q.SetTable(j.TableName())
	return r.dbclient.Create(q.BaseData, j)
}

func (r *dbJobRepo) Get(code string) (job *model.Job, err error) {
	q := db.NewQuery()
	q.SetTable(model.Job{}.TableName())
	q.SetCondition("code = ?", code)
	j := model.Job{}
	err = r.dbclient.First(q, &j)
	return &j, err
}

func (r *dbJobRepo) RemoveJob(code string) error {
	q := db.NewQuery()
	q.SetTable(model.Job{}.TableName())
	q.SetCondition("code = ?", code)
	var total int64
	return r.dbclient.Remove(q.BaseData, &total)

}
func (r *dbJobRepo) UpdateJob(j *model.Job) error {
	q := db.NewQuery()
	q.SetTable(j.TableName())
	q.SetCondition("code = ?", j.Code)
	var total int64
	commomMap := db.CommonMap{
		"status": j.Status,
	}
	if j.Cron != "" {
		commomMap["cron"] = j.Cron
	}
	if j.Title != "" {
		commomMap["title"] = j.Title
	}
	if j.Remark != "" {
		commomMap["remark"] = j.Remark
	}
	return r.dbclient.Updates(q.BaseData, commomMap, &total)
}

func (r *dbJobRepo) StartJob(code string) (err error) {
	updates := db.CommonMap{
		"status": 1,
	}
	q := db.NewQuery()
	q.SetTable(model.Job{}.TableName())
	q.SetCondition("code = ?", code)
	var rows int64
	return r.dbclient.Updates(q.BaseData, updates, &rows)
}

func (r *dbJobRepo) PauseJob(code string) (err error) {
	updates := db.CommonMap{
		"status": 0,
	}
	q := db.NewQuery()
	q.SetTable(model.Job{}.TableName())
	q.SetCondition("code = ?", code)
	var rows int64
	return r.dbclient.Updates(q.BaseData, updates, &rows)
}

func (r *dbJobRepo) CreateTask(j *model.Job) (*model.Task, error) {
	t := &model.Task{}
	t.Code = j.Code
	t.URL = j.URL
	t.StartAt = time.Now()
	q := db.NewQuery()
	q.SetTable(t.TableName())
	err := r.dbclient.Create(q.BaseData, t)
	return t, err
}

func (r *dbJobRepo) FeedbackTask(taskid uint, errno int, data interface{}) (err error) {
	updates := db.CommonMap{
		"status": errno,
		"end_at": time.Now(),
		"log":    fmt.Sprintf("%v", data),
	}
	q := db.NewQuery()
	q.SetTable(model.Task{}.TableName())
	q.SetCondition("id = ?", taskid)
	var rows int64
	return r.dbclient.Updates(q.BaseData, updates, &rows)
}
