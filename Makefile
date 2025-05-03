BIN?=html-analyzer
REGISTRY?=localhost
TAG?=latest
GIT_SSH_KEY?=~/.ssh/id_rsa
PROJECT_ROOT=$(shell pwd)

default: run
.PHONY : build run fresh test clean build-static docker-build docker-release

lint:
	golangci-lint run -c .golangci.yml --sort-results

test:
	GO111MODULE=on GOPRIVATE="github.com" go test ./... -tags musl -coverprofile=coverage.txt -covermode count

test-dynamic:
	GO111MODULE=on GOPRIVATE="github.com" go test ./... -tags=dynamic,musl --cover

test-show-coverage:
    GO111MODULE=on GOPRIVATE="github.com" go test -coverprofile=test_coverage/coverage.out ./internal/app/services/ && go tool cover -html=test_coverage/coverage.out -o test_coverage/coverage.html

build:
	GO111MODULE=on GOPRIVATE="github.com" go build -o build/${BIN}

build-mocks:
	cd mocks/ && rm -rf -- */ && mockery --all

docker-build:
	docker build --no-cache --build-arg BIN=${BIN} --build-arg GIT_SSH_KEY="$$(cat $(GIT_SSH_KEY))" -t ${BIN} .
	docker tag ${BIN} ${REGISTRY}/${BIN}

docker-ecr:
	docker build -f Dockerfile.ECR --build-arg BIN_PATH=./build/${BIN} -t ${BIN}:${TAG} .
	docker tag ${BIN}:${TAG} ${REGISTRY}/${REPOSITORY_NAME}:${TAG}
	docker push ${REGISTRY}/${REPOSITORY_NAME}:${TAG}

deploy:
	export VERSION=$(cat version.txt)
	curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
	sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
	aws eks update-kubeconfig --name ${EKS_CLUSTER_NAME} --region ${AWS_DEFAULT_REGION} --role-arn ${EKS_ROLE_ARN}
	curl https://get.helm.sh/helm-v3.8.2-linux-amd64.tar.gz --output helm-v3.8.2-linux-amd64.tar.gz
	sudo tar -zxvf helm-v3.8.2-linux-amd64.tar.gz
	sudo mv linux-amd64/helm /usr/local/bin/helm
	export HELM_EXPERIMENTAL_OCI=1
	aws ecr get-login-password --region ${AWS_ECR_REGION} | helm registry login --username AWS --password-stdin ${AWS_ECR_ACCOUNT_ID}.dkr.ecr.${AWS_ECR_REGION}.amazonaws.com
	envsubst < ./deployment/helm/values.yaml.template > ./deployment/helm/values.yaml
	envsubst < ./deployment/helm/Chart.yaml.template > ./deployment/helm/Chart.yaml
	helm dependency update ./deployment/helm
	helm upgrade ${ENV_NAME}-${CHART_NAME} ./deployment/helm -f ./deployment/helm/values.yaml --install --namespace bes --create-namespace --timeout 10m30s

run: build
	./build/${BIN}

clean:
	go clean
	rm -rf build
