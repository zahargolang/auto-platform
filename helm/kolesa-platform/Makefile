RELEASE := auto-platform
NAMESPACE := auto-platform
VALUES := -f values.yaml -f values-minikube.yaml

first-deploy: ## Первый деплой чарта в minikube
	helm install $(RELEASE) . -n $(NAMESPACE) --create-namespace $(VALUES)

deploy: 
	helm install $(RELEASE) . -n $(NAMESPACE) -f values.yaml -f values-minikube.yaml

redeploy: ## Передеплой (helm upgrade) после правок в чарте/сервисах
	helm upgrade $(RELEASE) . -n $(NAMESPACE) $(VALUES)

undeploy: ## андэплой чарта в minikube
	helm uninstall $(RELEASE) -n $(NAMESPACE)



snippet:
	kubectl patch configmap ingress-nginx-controller -n ingress-nginx \
  --type merge -p '{"data":{"allow-snippet-annotations":"true","annotations-risk-level":"Critical"}}'
	kubectl rollout restart deployment ingress-nginx-controller -n ingress-nginx
	kubectl rollout status deployment ingress-nginx-controller -n ingress-nginx