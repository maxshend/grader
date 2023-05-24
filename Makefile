grader_up:
	docker compose -f './deployments/docker-compose.yaml' build web worker runner && \
	docker compose -f './deployments/docker-compose.yaml' up
grader_web_up:
	docker compose -f './deployments/docker-compose.yaml' build web && \
	docker compose -f './deployments/docker-compose.yaml' up web
grader_down:
	docker compose -f './deployments/docker-compose.yaml' down -v
grader_logs:
	docker compose -f './deployments/docker-compose.yaml' logs --tail 10 --follow
grader_postgres:
	docker compose -f './deployments/docker-compose.yaml' exec -it postgres psql -U postgres -d grader

.PHONY: grader_up grader_down grader_logs grader_web_up grader_postgres
