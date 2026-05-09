.PHONY : api-build mysql docker-build api-get

api-build:
	goctl api go -api api/main.api -dir .

mysql:
    docker compose exec mysql mysql -uroot -pyourpassword

docker-build:
	docker-compose up --build

api-get:
	goctl api swagger --api api/main.api --dir docs