package repo

import (
	"fmt"
	"time"

	"github.com/team4yf/fpm-go-plugin-cron/model"
	"github.com/team4yf/yf-fpm-server-go/fpm"
)

//JobRepo the repo about the job
type JobRepo interface {
	List() ([]*model.Job, error)
	CreateJob(*model.Job) error
	StartJob(code string) error
	PauseJob(code string) error

	CreateTask(*model.Job) (*model.Task, error)
	FeedbackTask(taskid uint, errno int, data interface{}) error
	Tasks(code string, skip, limit int) ([]*model.Task, int, error)
}

type memoryJobRepo struct {
	jobs  map[string]*model.Job
	tasks map[uint]*model.Task
	seq   uint
}

func (r *memoryJobRepo) List() ([]*model.Job, error) {
	list := make([]*model.Job, 0)
	for _, j := range r.jobs {
		list = append(list, j)
	}
	return list, nil
}

func (r *memoryJobRepo) Tasks(code string, skip, limit int) ([]*model.Task, int, error) {
	list := make([]*model.Task, 0)
	for _, t := range r.tasks {
		if t.Code == code {
			list = append(list, t)
		}
	}
	total := len(list)
	list = list[skip : skip+limit]
	return list, total, nil
}

func (r *memoryJobRepo) CreateJob(j *model.Job) (err error) {
	r.jobs[j.Code] = j
	return
}

func (r *memoryJobRepo) StartJob(code string) (err error) {
	r.jobs[code].Status = 1
	return
}

func (r *memoryJobRepo) PauseJob(code string) (err error) {
	r.jobs[code].Status = 0
	return
}

func (r *memoryJobRepo) CreateTask(j *model.Job) (*model.Task, error) {
	t := &model.Task{}
	r.seq++
	t.ID = r.seq
	t.Code = j.Code
	t.URL = j.URL
	t.StartAt = time.Now()
	r.tasks[t.ID] = t
	return t, nil
}

func (r *memoryJobRepo) FeedbackTask(taskid uint, errno int, data interface{}) (err error) {
	t := r.tasks[taskid]
	t.Status = errno
	t.EndAt = time.Now()
	t.Cost = t.EndAt.UnixNano()/1e6 - t.StartAt.UnixNano()/1e6
	t.Log = fmt.Sprintf("%v", data)
	return
}

//NewRepo create a new job repo
func NewRepo(store string) JobRepo {
	switch store {
	case "memory":
		return &memoryJobRepo{
			jobs:  make(map[string]*model.Job),
			tasks: make(map[uint]*model.Task),
		}
	case "disk":
		return &diskJobRepo{
			jobs:  make(map[string]*model.Job),
			tasks: make(map[uint]*model.Task),
		}
	case "config":
		// read only
		if !fpm.Default().HasConfig("jobs") {
			panic("should insert [jobs] node in the config file")
		}
		jobs := make(map[string]*model.Job)
		if err := fpm.Default().FetchConfig("jobs", &jobs); err != nil {
			panic(fmt.Errorf("fetch [jobs] error: %v", err))
		}
		return &configJobRepo{
			jobs:  jobs,
			tasks: make(map[uint]*model.Task),
		}
	case "db":
		dbclient, ok := fpm.Default().GetDatabase("pg")
		if !ok {
			panic(fmt.Errorf("database plugin should installed! but not found"))
		}
		migrator := model.Migration{
			DS: dbclient,
		}
		migrator.Install()
		return &dbJobRepo{
			dbclient: dbclient,
		}
	}

	return nil
}
