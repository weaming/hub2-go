build-release:
	GOARCH=amd64 GOOS=linux go build -ldflags '-s -w' -o hub2-go-linux .
	# GOARCH=amd64 GOOS=darwin go build -ldflags '-s -w' -o hub2-go-darwin .