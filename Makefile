BINARY_NAME=graph
mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
mkfile_dir := $(dir $(mkfile_path))

all: docker-build docker-push

build:
	go build -o ${BINARY_NAME} main.go

run:
	./${BINARY_NAME} daemon

build_and_run: build run daemon

protob:
	protoc --go_out=. --go-grpc_out=.  pkg/daemon/daemon.proto
	protoc --go_out=.  pkg/graph/data.proto

generate:
	go generate ./...

clean:
	go clean
	rm ${BINARY_NAME}

docker-build:
	docker build --tag graph --tag "gcr.io/glif-292320/graph:$$(git rev-parse HEAD)" .

docker-push:
	docker push "gcr.io/glif-292320/graph:$$(git rev-parse HEAD)"

namespace:
	kubectl create namespace glif

update-calibration-env:
	kubectl delete secret calibration-graph-credentials -n glif
	kubectl create secret generic calibration-graph-credentials --from-env-file ./calibration.env -n glif

update-wallaby-env:
	kubectl delete secret wallaby-graph-credentials -n glif
	kubectl create secret generic wallaby-graph-credentials --from-env-file ./wallaby.env -n glif

update-mainnet-env:
	kubectl delete secret mainnet-graph-credentials -n glif
	kubectl create secret generic mainnet-graph-credentials --from-env-file ./mainnet.env -n glif

ingress:
	cd ./k8s/base && kubectl apply -f ingress.yml -n glif

reset-wallaby:
	kubectl -n glif scale --replicas=0 deployment/wallaby-graph-deployment
	kubectl -n glif delete pvc pvc-wallaby-graph
	kubectl -n glif apply -f k8s/wallaby/pvc.yml
	kubectl -n glif scale --replicas=1 deployment/wallaby-graph-deployment
	kubectl -n glif rollout status deployment/wallaby-graph-deployment
	kubectl -n glif get pods -l app=wallaby-graph-deployment -o name | xargs -I{} kubectl -n glif exec {} -- /app/giraph sync

service-calibration:
	cd ./k8s/calibration && kubectl apply -f service.yml -n glif
	cd ./k8s/calibration && kubectl apply -f certificate.yml -n glif

service-wallaby:
	cd ./k8s/wallaby && kubectl apply -f service.yml -n glif
	cd ./k8s/wallaby && kubectl apply -f certificate.yml -n glif

service-mainnet:
	cd ./k8s/mainnet && kubectl apply -f service.yml -n glif
	cd ./k8s/mainnet && kubectl apply -f certificate.yml -n glif

deploy-calibration:
	cd ./k8s/calibration && kustomize edit set image "gcr.io/PROJECT_ID/IMAGE:TAG=gcr.io/glif-292320/graph:$$(git rev-parse HEAD)"
	cd ./k8s/calibration && kustomize build . | kubectl apply -n glif -f -
	kubectl rollout status deployment/calibration-graph-deployment -n glif

deploy-wallaby:
	cd ./k8s/wallaby && kustomize edit set image "gcr.io/PROJECT_ID/IMAGE:TAG=gcr.io/glif-292320/graph:$$(git rev-parse HEAD)"
	cd ./k8s/wallaby && kustomize build . | kubectl apply -n glif -f -
	kubectl rollout status deployment/wallaby-graph-deployment -n glif

deploy-mainnet:
	cd ./k8s/mainnet && kustomize edit set image "gcr.io/PROJECT_ID/IMAGE:TAG=gcr.io/glif-292320/graph:$$(git rev-parse HEAD)"
	cd ./k8s/mainnet && kustomize build . | kubectl apply -n glif -f -
	kubectl rollout status deployment/mainnet-graph-deployment -n glif
