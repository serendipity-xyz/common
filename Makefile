
#  docker run -it --entrypoint /bin/sh serendipity-common-tester
test:
	docker build -t serendipity-common-tester .
	docker run --rm serendipity-common-tester
