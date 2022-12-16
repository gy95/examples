# make light
.PHONY: default image

default: image

image: bin
	docker build -t guoyao95/light:v1.0.0 .
	docker push guoyao95/light:v1.0.0

.PHONY: bin
bin:
	go build -o light

.PHONY: run
run: bin
	./light

