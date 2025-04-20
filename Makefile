runDocker:
	docker build -t booking-service .
	docker run --env-file .env -p 8081:8081 booking-service