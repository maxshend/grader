grader_web_up:
	docker compose -f './deployments/docker-compose.yaml' build web && \
	docker compose -f './deployments/docker-compose.yaml' up web
docker_down:
	docker compose -f './deployments/docker-compose.yaml' down -v

.PHONY: grader_web_up
