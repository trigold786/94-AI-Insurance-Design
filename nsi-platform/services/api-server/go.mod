module github.com/trigold786/94-AI-Insurance-Design/api-server

go 1.24

require (
	github.com/lib/pq v1.10.9
	github.com/trigold786/94-AI-Insurance-Design/shared v0.0.0
)

require github.com/golang-jwt/jwt/v5 v5.3.1 // indirect

replace github.com/trigold786/94-AI-Insurance-Design/shared => ../../shared
