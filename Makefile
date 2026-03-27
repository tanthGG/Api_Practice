# Tidy dependencies
tidy:
	cd Backend && go mod tidy

# Run the application
run:
	cd Backend && go run ./cmd/main.go

# Start docker-compose stack
compose-up:
	docker compose up -d

# Stop docker-compose stack
compose-down:
	docker compose down
