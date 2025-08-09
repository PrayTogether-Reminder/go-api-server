.PHONY: help build run test clean migrate docker-build docker-run

# 변수 정의
APP_NAME=go-api-server
MAIN_PATH=cmd/api/main.go
DOCKER_IMAGE=pray-together-api
DOCKER_TAG=latest

## help: 사용 가능한 명령어 표시
help:
	@echo 'Usage:'
	@echo '  make <command>'
	@echo ''
	@echo 'Commands:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## build: 애플리케이션 빌드
build:
	@echo 'Building ${APP_NAME}...'
	@go build -o bin/${APP_NAME} ${MAIN_PATH}
	@echo 'Build complete!'

## run: 애플리케이션 실행
run:
	@echo 'Running ${APP_NAME}...'
	@go run ${MAIN_PATH}

## test: 테스트 실행
test:
	@echo 'Running tests...'
	@go test -v ./...

## test-coverage: 테스트 커버리지 확인
test-coverage:
	@echo 'Running tests with coverage...'
	@go test -v -cover ./...

## clean: 빌드 파일 제거
clean:
	@echo 'Cleaning...'
	@rm -rf bin/
	@go clean
	@echo 'Clean complete!'

## deps: 의존성 다운로드
deps:
	@echo 'Downloading dependencies...'
	@go mod download
	@go mod tidy
	@echo 'Dependencies downloaded!'

## migrate: 데이터베이스 마이그레이션 실행
migrate:
	@echo 'Running database migrations...'
	@go run ${MAIN_PATH} migrate

## docker-build: Docker 이미지 빌드
docker-build:
	@echo 'Building Docker image...'
	@docker build -t ${DOCKER_IMAGE}:${DOCKER_TAG} .
	@echo 'Docker image built!'

## docker-run: Docker 컨테이너 실행
docker-run:
	@echo 'Running Docker container...'
	@docker run -p 8080:8080 --env-file .env ${DOCKER_IMAGE}:${DOCKER_TAG}

## swagger: Swagger 문서 생성
swagger:
	@echo 'Generating Swagger documentation...'
	@swag init -g ${MAIN_PATH}
	@echo 'Swagger documentation generated!'

## lint: 코드 린트 실행
lint:
	@echo 'Running linter...'
	@golangci-lint run ./...
	@echo 'Linting complete!'

## fmt: 코드 포맷팅
fmt:
	@echo 'Formatting code...'
	@go fmt ./...
	@echo 'Formatting complete!'
