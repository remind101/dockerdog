bin/dockerdog: *.go
	docker build -t remind101/dockerdog .
	docker cp $(shell docker create remind101/dockerdog):/go/bin/dockerdog bin/
