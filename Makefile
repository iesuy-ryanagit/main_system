# 環境変数
NAMESPACE = default

# イメージ名
FRONTEND_IMAGE = ryanagit/react-app:v1
APP_IMAGE = ryanagit/demo:v2

# 各種リソースファイル
FRONTEND_DEPLOYMENT = frontend-deployment.yaml
FRONTEND_SERVICE = frontend-service.yaml

APP_DEPLOYMENT = go-app-deployment.yaml
APP_SERVICE = go-app-service.yaml

# イメージ名（MySQLは公式イメージを使う場合）
MYSQL_IMAGE = mysql:8.0

# build: frontend
build-frontend:
	@echo "Building Docker image for frontend..."
	docker build -t $(FRONTEND_IMAGE) --no-cache frontend

# build: app
build-app:
	@echo "Building Docker image for app..."
	docker build -t $(APP_IMAGE) --no-cache app

# build: 両方まとめて
build-all: build-frontend build-app
	@echo "✅ All builds complete."

# deploy frontend
deploy-frontend:
	@echo "🚀 Deploying frontend..."
	kubectl set image deployment/frontend-deployment react-app=$(FRONTEND_IMAGE) -n $(NAMESPACE)
	kubectl rollout restart deployment/frontend-deployment -n $(NAMESPACE)

# deploy app
deploy-app:
	@echo "🚀 Deploying app..."
	kubectl set image deployment/go-app go-app=$(APP_IMAGE) -n $(NAMESPACE)
	kubectl rollout restart deployment/go-app -n $(NAMESPACE)

# deploy mysql (イメージ更新は不要なのでapplyだけでもOK)
deploy-mysql:
	@echo "🚀 Applying MySQL manifests..."
	kubectl apply -f mysql-deployment.yaml -n $(NAMESPACE)
	kubectl apply -f mysql-service.yaml -n $(NAMESPACE)

# deploy all
deploy-all: build-all deploy-frontend deploy-app deploy-mysql
	@echo "✅ All deploy complete."

# apply manifests (frontend, app, mysql)
apply:
	kubectl apply -f frontend-deployment.yaml -n $(NAMESPACE)
	kubectl apply -f frontend-service.yaml -n $(NAMESPACE)
	kubectl apply -f go-app-deployment.yaml -n $(NAMESPACE)
	kubectl apply -f go-app-service.yaml -n $(NAMESPACE)
	kubectl apply -f mysql-deployment.yaml -n $(NAMESPACE)
	kubectl apply -f mysql-service.yaml -n $(NAMESPACE)

# logs
logs-frontend:
	kubectl logs -f $$(kubectl get pods -l app=frontend -o jsonpath='{.items[0].metadata.name}')

logs-app:
	kubectl logs -f $$(kubectl get pods -l app=go-app -o jsonpath='{.items[0].metadata.name}')

logs-mysql:
	kubectl logs -f $$(kubectl get pods -l app=mysql -o jsonpath='{.items[0].metadata.name}')

# status
status:
	kubectl get all -n $(NAMESPACE)
