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
    store: memory    # memory is the default, support : config, db
```

### Config via yml
```yaml
cron:
  store: config

jobs:
  test:
    name: "test"
    code: "test"
    cron: "* * * * *"
    status: 1
    executeType: "INTERNAL"
    url: "foo.bar"
    timeout: 600
    retryMax: 2
    notifyTopic: "test"
### Topics

- #job/done, and payload is
  - event normal its the job.topic
  - errno 0 is ok, the other is error
  - body the data or error message

