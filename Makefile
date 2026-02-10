build: generate
	go build .

generate:
	npx buf generate

docker-build:
	 docker build --build-arg SERVICE_NAME=cottage-manager --build-arg VERSION=latest -t cottage-manager:latest .