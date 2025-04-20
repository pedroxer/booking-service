runDocker:
	docker build -t booking-service .
	docker run --env-file .env -p 8080:8080 booking-service