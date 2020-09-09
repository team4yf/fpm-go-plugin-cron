package plugin

import (
	"github.com/team4yf/fpm-go-plugin-cron/job"
	"github.com/team4yf/fpm-go-plugin-cron/model"
	"github.com/team4yf/fpm-go-plugin-cron/repo"
	"github.com/team4yf/yf-fpm-server-go/fpm"
)

type cronConfig struct {
	Store string
}

type codeReq struct {
	Code  string  `json:"code"`
	Skip  float64 `json:"skip,omitempty"`
	Limit float64 `json:"limit,omitempty"`
}

func init() {
	var jobService job.JobService
	fpm.RegisterByPlugin(&fpm.Plugin{
		Name: "fpm-plugin-cron",
		V:    "0.0.1",
		Handler: func(fpmApp *fpm.Fpm) {
			fpmApp.AddHook("AFTER_INIT", func(_ *fpm.Fpm) {
				// fetch config
				cronSetting := cronConfig{
					Store: "memory",
				}
				if fpmApp.HasConfig("cron") {
					if err := fpmApp.FetchConfig("cron", &cronSetting); err != nil {
						panic(err)
					}
				}

				jobService = job.NewSimpleJobService(repo.NewRepo(cronSetting.Store))
				jobService.Init()
				jobService.Start()
				fpmApp.AddBizModule("job", &fpm.BizModule{
					"list": func(param *fpm.BizParam) (data interface{}, err error) {
						data, err = jobService.List()
						return
					},
					"add": func(param *fpm.BizParam) (data interface{}, err error) {
						var req model.Job
						if err = param.Convert(&req); err != nil {
							return
						}
						err = jobService.Add(&req)
						data = 1
						return
					},
					"execute": func(param *fpm.BizParam) (data interface{}, err error) {
						var req codeReq
						if err = param.Convert(&req); err != nil {
							return
						}
						data, err = jobService.Execute(req.Code)
						return
					},
					"restart": func(param *fpm.BizParam) (data interface{}, err error) {
						var req codeReq
						if err = param.Convert(&req); err != nil {
							return
						}
						err = jobService.Restart(req.Code)
						data = 1
						return
					},
					"pause": func(param *fpm.BizParam) (data interface{}, err error) {
						var req codeReq
						if err = param.Convert(&req); err != nil {
							return
						}
						err = jobService.Pause(req.Code)
						data = 1
						return
					},
					"tasks": func(param *fpm.BizParam) (data interface{}, err error) {
						var req codeReq
						if err = param.Convert(&req); err != nil {
							return
						}
						list, total, err := jobService.Tasks(req.Code, int(req.Skip), int(req.Limit))
						return map[string]interface{}{
							"row":   list,
							"total": total,
						}, err
					},
				})
			}, 1)

			// fpmApp.Subscribe("#job/done", func(topic string, payload interface{}) {
			// 	fpmApp.Logger.Debugf("t %s, p %v", topic, payload)
			// })
		},
	})
}
