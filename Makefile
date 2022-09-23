
test:
	docker build -t serendipity-core-tester .
	docker run --rm serendipity-core-tester