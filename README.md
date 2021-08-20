## Requirements

### Framework

$$
Client \stackrel{HTTP}{\rightarrow} Server(HTTP) \stackrel{TCP}{\rightarrow} Server(TCP) \stackrel{DB}{\rightarrow} MySQL/Redis
$$

### User Practice in Client

- Login
- Show profile
- Change nickname
- Update avatar

### HTTP server

- Deliver requests between front-end and TCP server

### TCP server

Major logic (MySQL, Redis)

## Implementation Logic

When it comes to the high concurrent requests, it is neccessary to cache the account's infomation with the combination of MySQL and Redis.

### consistency bewtween mysql and redis

Considering the consistency issue regarding updating the data, I delete the cache before update in the db in case of inconsistency.

## References

- Go: http://golang.org
- Coding style: https://github.com/golang/go/wiki/CodeReviewComments
- Testing: https://golang.org/pkg/testing/
- Profiling: http://blog.golang.org/profiling-go-programs
- Go Web application example: https://golang.org/doc/articles/wiki/
- Go editor/IDE
  - https://github.com/fatih/vim-go
  - https://github.com/dominikh/go-mode.el
  - https://github.com/DisposaBoy/GoSublime
  - https://github.com/visualfc/liteide
- MySQL client: https://github.com/go-sql-driver/mysql
- Redis: http://redis.io
- Redis Client: https://github.com/go-redis/redis

` go run main.go &
wrk -t4 -c800 -s ./test/mulTest.lua -d10s http://127.0.0.1:5500/`
