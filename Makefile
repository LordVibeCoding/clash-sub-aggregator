.PHONY: build run clean web

build: web
	go build -o bin/clash-sub-aggregator .

web:
	cd web && npm install && npm run build

run: build
	./bin/clash-sub-aggregator

docker-build:
	docker compose build

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

clean:
	rm -rf bin/ static/
