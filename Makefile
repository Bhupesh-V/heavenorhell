APP_NAME=heavenorhell-prod

linux:
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(APP_NAME)
