# go-fpm-plugin-cron


a cron plugin for fpm-server, it will support `INTERNAL` & `WEB` actions.

the INTERNAL can invoke the local method.

the WEB can send a http request.

### Install

`$ go get -u github.com/team4yf/fpm-go-plugin-cron`


import: 
```golang
import _ "github.com/team4yf/fpm-go-plugin-cron/plugin"
```

### Config
```yaml
cron:
    store: memory    # memory is the default
```

### Topics

- #job/done, and payload is
  - event normal its the job.topic
  - errno 0 is ok, the other is error
  - body the data or error message

