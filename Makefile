# Tidy dependencies
tidy:
	cd Backend && go mod tidy

	# Run the application
run:
	cd Backend && go run ./cmd/main.go
