module github.com/onurbilginnn/gator

replace github.com/onurbilginnn/internal/config => ./internal/config

go 1.26.4

require github.com/onurbilginnn/internal/config v0.0.0

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/lib/pq v1.12.3 // indirect
)
