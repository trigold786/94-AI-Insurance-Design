module github.com/trigold786/94-AI-Insurance-Design/api-server

go 1.24

require (
	github.com/gin-gonic/gin v1.10.0
	github.com/lib/pq v1.10.9
	github.com/redis/go-redis/v9 v9.3.0
	github.com/trigold786/94-AI-Insurance-Design/shared v0.0.0
)

replace github.com/trigold786/94-AI-Insurance-Design/shared => ../../shared
