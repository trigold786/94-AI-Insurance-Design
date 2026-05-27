module github.com/trigold786/94-AI-Insurance-Design/policy-crawler

go 1.25.0

require (
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/lib/pq v1.10.9
	github.com/pgvector/pgvector-go v0.4.0
	github.com/trigold786/94-AI-Insurance-Design/shared v0.0.0
)

require (
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/ledongthuc/pdf v0.0.0-20250511090121-5959a4027728 // indirect
)

replace github.com/trigold786/94-AI-Insurance-Design/shared => ../../shared
