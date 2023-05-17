grader_up:
	docker compose -f './deployments/docker-compose.yaml' build web worker && \
	docker compose -f './deployments/docker-compose.yaml' up
docker_down:
	docker compose -f './deployments/docker-compose.yaml' down -v
grader_logs:
	docker compose -f './deployments/docker-compose.yaml' logs --tail 10 --follow

.PHONY: grader_web_up
