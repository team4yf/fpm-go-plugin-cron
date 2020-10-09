package plugin

import (
	"fmt"
	"strconv"

	"github.com/team4yf/fpm-go-plugin-cron/job"
	"github.com/team4yf/fpm-go-plugin-cron/model"
	"github.com/team4yf/fpm-go-plugin-cron/repo"
	"github.com/team4yf/yf-fpm-server-go/fpm"
)

type cronConfig struct {
	Store string
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
						return
					},
					"update": func(param *fpm.BizParam) (data interface{}, err error) {
						var req model.Job
						if err = param.Convert(&req); err != nil {
							return
						}
						err = jobService.Update(&req)
						return
					},
					"execute": func(param *fpm.BizParam) (data interface{}, err error) {
						code := (*param)["code"].(string)
						data, err = jobService.Execute(code)
						return
					},
					"restart": func(param *fpm.BizParam) (data interface{}, err error) {
						code := (*param)["code"].(string)
						err = jobService.Restart(code)
						return
					},
					"get": func(param *fpm.BizParam) (data interface{}, err error) {
						code := (*param)["code"].(string)
						data, err = jobService.Get(code)
						return
					},
					"pause": func(param *fpm.BizParam) (data interface{}, err error) {
						code := (*param)["code"].(string)
						err = jobService.Pause(code)
						return
					},
					"remove": func(param *fpm.BizParam) (data interface{}, err error) {
						code := (*param)["code"].(string)
						err = jobService.Remove(code)
						return
					},
					"tasks": func(param *fpm.BizParam) (data interface{}, err error) {
						code := (*param)["code"].(string)
						limitNum := -1
						skipNum := 0
						skip, ex := (*param)["skip"]
						if ex {
							skipNum, err = strconv.Atoi(fmt.Sprintf("%v", skip))
						}
						limit, ex := (*param)["limit"]
						if ex {
							limitNum, err = strconv.Atoi(fmt.Sprintf("%v", limit))
						}
						list, total, err := jobService.Tasks(code, skipNum, limitNum)
						return map[string]interface{}{
							"row":   list,
							"total": total,
						}, err
					},
				})
			}, 1)
		},
	})
}
