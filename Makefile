grader_web_up:
	docker compose -f './deployments/docker-compose.yaml' build web && \
	docker compose -f './deployments/docker-compose.yaml' up web

.PHONY: grader_web_up
