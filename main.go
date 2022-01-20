package main

import (
	// // _ "github.com/team4yf/fpm-go-plugin-cron/plugin"
	// _ "github.com/team4yf/fpm-go-plugin-orm/plugins/pg"
	_ "github.com/team4yf/fpm-go-plugin-cron/plugin"
	"github.com/team4yf/yf-fpm-server-go/fpm"
)

func main() {

	fpmApp := fpm.New()
	fpmApp.Init()
	fpmApp.Run()
}
