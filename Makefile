
# required for 'make push'
DOCKER_IMAGE ?= 462307740163.dkr.ecr.us-east-2.amazonaws.com/scribble
DOCKER_TAG ?= latest


# bundle www src
bundle:
	cd www && yarn install && npm run-script build

# package resources for Go
pkged.go: resources/game.js ./resources/* ./templates/*
	rm -f pgked.go
	go run github.com/markbates/pkger/cmd/pkger -include /resources -include /templates;

# clean generated files
clean:
	rm -f pkged.go
	rm -f scribblers
	rm -f resources/game.js
	rm -f resources/game.js.map

# build Go binary
build: pkged.go
	GO111MODULE=on CGO_ENABLED=0 go build -o scribblers

# build docker image with go binary
dockerize:
	docker build -t ${DOCKER_IMAGE}:${DOCKER_TAG} .
	docker tag ${DOCKER_IMAGE}:${DOCKER_TAG} scribble:latest

# push docker image to repo
push:
	docker push ${DOCKER_IMAGE}:${DOCKER_TAG}	

# run locally -- first 
up:
	docker-compose up -d scribblers redis

log:
	docker-compose logs -f scribblers

# stop locally
down: 
	docker-compose down;

# destroy local
destroy: down
	docker-compose rm -f;

# start / bounce / stop redis for tests
bounce-redis:
	docker-compose exec redis_test redis-cli flushall

start-redis:
	docker-compose up -d redis_test
	docker-compose exec redis_test redis-cli flushall
	
stop-redis:
	docker-compose stop
	docker-compose rm -f

# run all tests
test: bounce-redis
	go test -v ./...

# run tests for game package
test-game: bounce-redis
	go test -v ./game

PHONY: test, stop-redis start-redis bounce-redis destroy down up dockerize build clean