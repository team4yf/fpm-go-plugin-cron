package job

import (
	"fmt"
	"testing"

	"github.com/team4yf/fpm-go-plugin-cron/model"
	"github.com/team4yf/fpm-go-plugin-cron/repo"
	_ "github.com/team4yf/fpm-go-plugin-orm/plugins/pg"
	"github.com/team4yf/yf-fpm-server-go/fpm"
)

func TestJobService(t *testing.T) {
	fpmApp := fpm.New()
	fpmApp.Init()

	jobService := NewSimpleJobService(repo.NewRepo("memory"))
	jobService.Init()
	jobService.Start()
	err := jobService.Add(&model.Job{
		Cron:        "* * * * *",
		Code:        "test",
		ExecuteType: "INTERNAL",
		Status:      1,
		URL:         "foo.bar",
		Argument:    "{}",
	})
	fmt.Printf("err: %v", err)
	data, err := jobService.Execute("test")
	fmt.Printf("data: %v, err: %v", data, err)

	list, err := jobService.List()
	fmt.Printf("list: %v, err: %v", list, err)
}

func TestJobService1(t *testing.T) {
	fpmApp := fpm.New()
	fpmApp.Init()

	jobService := NewSimpleJobService(repo.NewRepo("memory"))
	jobService.Init()
	jobService.Start()
	err := jobService.Add(&model.Job{
		Cron:        "* * * * *",
		Code:        "test",
		ExecuteType: "GET",
		Status:      1,
		URL:         "http://localhost:9090/health",
		Argument:    "{}",
	})
	fmt.Printf("err: %v", err)
	data, err := jobService.Execute("test")
	fmt.Printf("data: %v, err: %v", data, err)

	list, err := jobService.List()
	fmt.Printf("list: %v, err: %v", list, err)
}

func TestJobService2(t *testing.T) {
	fpmApp := fpm.New()
	fpmApp.Init()

	jobService := NewSimpleJobService(repo.NewRepo("db"))
	jobService.Init()
	jobService.Start()
	err := jobService.Add(&model.Job{
		Cron:        "* * * * *",
		Code:        "test",
		ExecuteType: "GET",
		Status:      1,
		URL:         "http://localhost:9090/health",
		Argument:    "{}",
	})
	err = jobService.Update(&model.Job{
		Cron:        "* * * * *",
		Code:        "test",
		ExecuteType: "GET",
		Status:      1,
		URL:         "http://localhost:9090/health",
		Argument:    "{}",
	})
	fmt.Printf("err: %v", err)
}
