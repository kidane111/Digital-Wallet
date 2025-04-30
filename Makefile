run:
	go run cmd/main.go
air:
	- go install github.com/air-verse/air@latest
swagger:
	-swag fmt && swag init -g cmd/main.go
test:
	- go test -v ./...
sqlc:
	cd ./config && sqlc generate
migrate-up:
	- migrate -path ./db/migrations/up -database cockroachdb://root:@localhost:26257/telematicsplatform?sslmode=disable  up
migrate-down:
	- migrate -path ./db/migrations/up -database cockroachdb://root:@localhost:26257/telematicsplatform?sslmode=disable down
