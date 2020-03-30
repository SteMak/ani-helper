BUILD_DIR=$(shell pwd)/bin
BOT_DIR=$(shell pwd)/worker
clean:
	rm -rf $(BUILD_DIR)/*

build:
	mkdir -p $(BOT_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64
	go build -v -o $(BUILD_DIR)/worker $(BOT_DIR)

docker-push: clean build
	heroku container:push worker -a ani-helper
	heroku container:release worker -a ani-helper