.PHONY : api-build mysql docker-build git-push

api-build:
	goctl api go -api api/main.api -dir .

mysql:
    docker compose exec mysql mysql -uroot -pyourpassword

docker-build:
	docker-compose up --build

