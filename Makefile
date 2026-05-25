.PHONY : api-build mysql docker-build api-get coverage coverage-func coverage-html

api-build:
	goctl api go -api api/main.api -dir . -style go_zero

mysql:
    docker compose exec mysql mysql -uroot -pyourpassword

docker-build:
	docker-compose up --build

api-get:
	goctl api swagger --api api/main.api --dir docs

coverage:
	go test ./... -coverprofile coverage.out

coverage-func:
	go tool cover -func coverage.out

coverage-html:
	go tool cover -html coverage.out -o coverage.html
