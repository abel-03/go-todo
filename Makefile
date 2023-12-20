build:
	cd frontend && npm run build
	go build -o ../bin
build-and-run:
	make build
	godotenv -f .env ./bin/go-todo
run:
	godotenv -f .env go run . & cd frontend && npm run start
deploy:
	cd frontend && npm run build
	gcloud app deploy
