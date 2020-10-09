package repo

import (
	"fmt"
	"time"

	"github.com/team4yf/fpm-go-plugin-cron/model"
)

type configJobRepo struct {
	jobs  map[string]*model.Job
	tasks map[uint]*model.Task
	seq   uint
}

func (r *configJobRepo) List() ([]*model.Job, error) {
	list := make([]*model.Job, 0)
	for _, j := range r.jobs {
		list = append(list, j)
	}
	return list, nil
}

func (r *configJobRepo) Tasks(code string, skip, limit int) ([]*model.Task, int, error) {
	list := make([]*model.Task, 0)
	for _, t := range r.tasks {
		if t.Code == code {
			list = append(list, t)
		}
	}
	total := len(list)
	if total >= skip+limit {
		list = list[skip : skip+limit]
	} else if skip < total {
		list = list[skip:total]
	}

	return list, total, nil
}

func (r *configJobRepo) CreateJob(j *model.Job) (err error) {
	return fmt.Errorf("CreateJob cant support the config Job Type")
}

func (r *configJobRepo) Get(code string) (*model.Job, error) {
	return r.jobs[code], nil
}

func (r *configJobRepo) RemoveJob(code string) error {
	delete(r.jobs, code)
	return nil
}
func (r *configJobRepo) UpdateJob(j *model.Job) error {
	r.jobs[j.Code] = j
	return nil
}

func (r *configJobRepo) StartJob(code string) (err error) {
	r.jobs[code].Status = 1
	return
}

func (r *configJobRepo) PauseJob(code string) (err error) {
	r.jobs[code].Status = 0
	return
}

func (r *configJobRepo) CreateTask(j *model.Job) (*model.Task, error) {
	t := &model.Task{}
	r.seq++
	t.ID = r.seq
	t.Code = j.Code
	t.URL = j.URL
	t.StartAt = time.Now()
	r.tasks[t.ID] = t
	return t, nil
}

func (r *configJobRepo) FeedbackTask(taskid uint, errno int, data interface{}) (err error) {
	t := r.tasks[taskid]
	t.Status = errno
	t.EndAt = time.Now()
	t.Cost = t.EndAt.UnixNano()/1e6 - t.StartAt.UnixNano()/1e6
	t.Log = fmt.Sprintf("%v", data)
	return
}
