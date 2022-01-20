package job

import (
	"errors"
	"fmt"
	"sync"

	"github.com/robfig/cron/v3"
	"github.com/team4yf/fpm-go-pkg/log"
	"github.com/team4yf/fpm-go-pkg/utils"
	"github.com/team4yf/fpm-go-plugin-cron/model"
	"github.com/team4yf/fpm-go-plugin-cron/repo"
	"github.com/team4yf/yf-fpm-server-go/fpm"
)

var (
	errNotInited     = errors.New("Schedule Not Inited")
	errJobCodeExists = errors.New("Job Code Exists")
	errJobNotExists  = errors.New("Job Not Exists")
	inited           = false //the flag of the service.Init()
)

//Init init by the caller
func Init() {
}

//Callback the callback struct
type Callback func(data interface{}, err error)

//JobService service for job
type JobService interface {
	Init() error
	List() ([]*model.Job, error)
	Start() error
	Add(job *model.Job) error
	Get(code string) (*model.Job, error)
	Update(job *model.Job) error
	Execute(code string) (interface{}, error)
	Restart(code string) error
	Pause(code string) error
	Remove(code string) error
	Tasks(code string, skip, limit int) ([]*model.Task, int, error)
	Shutdown() error
}

type jobWrapper struct {
	job *model.Job
	f   func()
	id  cron.EntryID
}
type simpleJobService struct {
	repo     repo.JobRepo
	schedule *cron.Cron
	locker   sync.RWMutex
	handler  map[string]*jobWrapper
}

//NewSimpleJobService create a new job service
func NewSimpleJobService(repo repo.JobRepo) JobService {
	serviec := &simpleJobService{
		repo:    repo,
		handler: make(map[string]*jobWrapper),
	}
	return serviec
}

func generateCallback(s *simpleJobService, theJob *model.Job) func() {
	return func() {
		go s.runJob(theJob, func(data interface{}, err error) {
			if err != nil {
				// log.Errorf("Run Job: %+v, Error: %+v\n", theJob, err)
				return
			}
		})
	}

}

func (s *simpleJobService) Init() (err error) {
	if inited {
		return nil
	}

	list, err := s.repo.List()
	if err != nil {
		return
	}
	s.locker.RLock()
	defer s.locker.RUnlock()
	for _, j := range list {
		s.handler[j.Code] = &jobWrapper{
			job: j,
			f:   generateCallback(s, j),
		}
	}
	s.schedule = cron.New()
	inited = true
	return nil
}

func (s *simpleJobService) List() ([]*model.Job, error) {
	return s.repo.List()
}

func (s *simpleJobService) Tasks(code string, skip, limit int) ([]*model.Task, int, error) {
	_, ok := s.handler[code]
	if !ok {
		//not exists
		return nil, 0, errors.New("job:" + code + ", not exists")
	}
	return s.repo.Tasks(code, skip, limit)
}

func (s *simpleJobService) Remove(code string) (err error) {
	//Stop
	if err = s.Pause(code); err != nil {
		return
	}
	delete(s.handler, code)
	//Remove
	return s.repo.RemoveJob(code)
}

func (s *simpleJobService) Get(code string) (*model.Job, error) {
	wrapper, ok := s.handler[code]
	if !ok {
		//not exists
		return nil, errors.New("job:" + code + ", not exists")
	}
	return wrapper.job, nil
}

func (s *simpleJobService) Update(job *model.Job) (err error) {
	if err = s.Pause(job.Code); err != nil {
		return
	}
	//update the job
	if err = s.repo.UpdateJob(job); err != nil {
		return
	}
	if job, err = s.repo.Get(job.Code); err != nil {
		return
	}
	//add the func
	wrapper := &jobWrapper{
		job: job,
		f:   generateCallback(s, job),
	}
	//AutoRun
	if wrapper.job.Status == 1 {
		id, err := s.schedule.AddFunc(wrapper.job.Cron, wrapper.f)
		if err != nil {
			return err
		}
		wrapper.id = id
	}

	s.handler[job.Code] = wrapper
	return nil
}

func (s *simpleJobService) Start() (err error) {
	if s.schedule == nil {
		return errNotInited
	}
	//add the func
	for _, wrapper := range s.handler {
		//ignore the not run job
		if wrapper.job.Status != 1 {
			continue
		}
		id, err := s.schedule.AddFunc(wrapper.job.Cron, wrapper.f)
		if err != nil {
			return err
		}
		wrapper.id = id
		log.Infof("Start() Job: Code-> %v; Corn-> %v;", wrapper.job.Code, wrapper.job.Cron)
	}
	//startup
	s.schedule.Start()
	return nil
}

func (s *simpleJobService) Add(j *model.Job) (err error) {
	if s.schedule == nil {
		return errNotInited
	}
	if _, ok := s.handler[j.Code]; ok {
		return errors.New("job:" + j.Code + ", exists")
	}
	if err = s.repo.CreateJob(j); err != nil {
		return
	}
	//add the func
	wrapper := &jobWrapper{
		job: j,
		f:   generateCallback(s, j),
	}
	//AutoRun
	if wrapper.job.Status == 1 {
		id, err := s.schedule.AddFunc(wrapper.job.Cron, wrapper.f)
		if err != nil {
			return err
		}
		wrapper.id = id
	}

	s.handler[j.Code] = wrapper
	return nil
}

func (s *simpleJobService) Execute(code string) (data interface{}, err error) {

	if _, ok := s.handler[code]; !ok {
		err = errJobNotExists
		return
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go s.runJob(s.handler[code].job, func(d interface{}, e error) {
		defer wg.Done()
		err = e
		data = d
	})
	wg.Wait()
	return
}

func (s *simpleJobService) Restart(code string) error {
	wrapper, ok := s.handler[code]
	if !ok {
		//not exists
		return errors.New("job:" + code + ", not exists")
	}

	if wrapper.id > 0 {
		//running
		return nil
	}
	if err := s.repo.StartJob(code); err != nil {
		return err
	}
	id, err := s.schedule.AddFunc(wrapper.job.Cron, wrapper.f)
	if err != nil {
		return err
	}
	wrapper.id = id
	return nil
}

func (s *simpleJobService) Pause(code string) error {
	wrapper, ok := s.handler[code]
	if !ok {
		//not exists
		return errors.New("job:" + code + ", not exists")
	}
	if wrapper.id < 0 {
		//need not to pause
		return nil
	}
	if err := s.repo.PauseJob(code); err != nil {
		return err
	}
	s.schedule.Remove(wrapper.id)
	wrapper.id = -99
	return nil
}

func (s *simpleJobService) Shutdown() error {
	s.schedule.Stop()
	s.schedule = nil
	return nil
}

//Actually run the http request
//Log the response for the task
func (s *simpleJobService) runJob(job *model.Job, callback Callback) {
	task, err := s.repo.CreateTask(job)
	if err != nil {
		callback(nil, err)
		return
	}

	errno := 0
	var body interface{}

	if job.ExecuteType == "INTERNAL" {
		param := fpm.BizParam{}
		if job.Argument != "" {
			if err := utils.StringToStruct(job.Argument, &param); err != nil {
				callback(nil, err)
				return
			}
		}
		rsp, err := fpm.Default().Execute(job.URL, &param, nil)
		if err != nil {
			errno = -1
			body = err.Error()
		} else {
			body = rsp
		}
	} else {
		var rsp utils.ResponseWrapper
		var auth *utils.HTTPAuth
		if job.Auth != "" {
			//construct the auth data
			authProp := job.AuthProperties

			auth = &utils.HTTPAuth{
				Type: utils.HTTPAuthType(job.Auth),
			}
			if authProp != "" && authProp != "{}" {
				utils.StringToStruct(authProp, &auth.Data)
			}
		}

		switch job.ExecuteType {
		case "POST":
			rsp = utils.PostJSONWithHeaderAndAuth(job.URL, []byte(job.Argument), job.Timeout, nil, auth)
		case "GET":
			rsp = utils.GetWithAuth(job.URL, job.Timeout, auth)
		case "FORM":
			rsp = utils.PostParamsWithHeaderAndAuth(job.URL, job.Argument, job.Timeout, nil, auth)

		default:
			body = "unsupported type:" + job.ExecuteType
			errno = -1
		}

		if errno != -1 {
			if rsp.StatusCode >= 200 && rsp.StatusCode <= 300 {
				errno = 0
				body = rsp.GetStringBody()
			} else {
				errno = -1
				body = rsp.Err
			}
		}
	}

	if err := s.repo.FeedbackTask(task.ID, errno, body); err != nil {
		callback(nil, err)
		return
	}
	fpm.Default().Publish("#job/done", map[string]interface{}{
		"event": job.NotifyTopic,
		"errno": errno,
		"body":  body,
	})
	if errno != 0 {
		callback(nil, fmt.Errorf("status not ok! %v", body))
		return
	}
	callback(body, nil)

}
