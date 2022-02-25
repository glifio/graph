all: build push

build:
	docker build --tag graph --tag "gcr.io/glif-292320/graph:$$(git rev-parse HEAD)" .

push:
	docker push "gcr.io/glif-292320/graph:$$(git rev-parse HEAD)"

namespace:
	kubectl create namespace glif

update-calibration-env:
	kubectl delete secret calibration-graph-credentials -n glif
	kubectl create secret generic calibration-graph-credentials --from-env-file ./calibration.env -n glif

update-mainnet-env:
	kubectl delete secret mainnet-graph-credentials -n glif
	kubectl create secret generic mainnet-graph-credentials --from-env-file ./calibration.env -n glif

ingress:
	cd ./k8s/base && kubectl apply -f ingress.yml -n glif

service-calibration:
	cd ./k8s/calibration && kubectl apply -f service.yml -n glif
	cd ./k8s/calibration && kubectl apply -f certificate.yml -n glif

service-mainnet:
	cd ./k8s/mainnet && kubectl apply -f service.yml -n glif
	cd ./k8s/mainnet && kubectl apply -f certificate.yml -n glif

deploy-calibration:
	cd ./k8s/calibration && kustomize edit set image "gcr.io/PROJECT_ID/IMAGE:TAG=gcr.io/glif-292320/graph:$$(git rev-parse HEAD)"
	cd ./k8s/calibration && kustomize build . | kubectl apply -n glif -f -
	kubectl rollout status deployment/calibration-graph-deployment -n glif

deploy-mainnet:
	cd ./k8s/mainnet && kustomize edit set image "gcr.io/PROJECT_ID/IMAGE:TAG=gcr.io/glif-292320/graph:$$(git rev-parse HEAD)"
	cd ./k8s/mainnet && kustomize build . | kubectl apply -n glif -f -
	kubectl rollout status deployment/mainnet-graph-deployment -n glif
