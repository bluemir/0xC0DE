deploy: build/docker-image.pushed
	# deploy code
	# cat deploy.yaml | DEPLOY_IMAGE=$(shell cat $<) envsubst | kubectl apply -f -

.PHONY: deploy
