.PHONY: build run clean

build:
	go build -o bin/clash-sub-aggregator .

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
	rm -rf bin/
