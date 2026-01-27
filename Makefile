# Tidy dependencies
tidy:
	go mod tidy

	# Run the application
run:
	go run ./cmd/main.go