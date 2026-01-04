run:
	npm run build --prefix frontend && go run .

run-docker:
	docker compose -f build/docker-compose.yaml up --build