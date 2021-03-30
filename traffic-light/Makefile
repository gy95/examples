# make traffic
.PHONY: default image

default: image

image: bin
	docker build -t kubeedge/traffic-light:update .

.PHONY: bin
bin:
	go build -o light

.PHONY: run
run: bin
	./light
