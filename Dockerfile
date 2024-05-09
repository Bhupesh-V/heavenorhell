FROM golang:alpine3.19
WORKDIR /src
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . ./
RUN go build -o heavenorhell main.go

CMD ["./heavenorhell"]