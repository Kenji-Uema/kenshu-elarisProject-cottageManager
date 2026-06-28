build: generate
	go build .

generate:
	npx buf generate

docker-build:
	 docker buildx build --build-arg SERVICE_NAME=cottage-manager --build-arg VERSION=1.0.2 -t cottage-manager:1.0.2 --load .