FROM golang:alpine3.19 as base
WORKDIR /src
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . ./
RUN go build -o heavenorhell main.go

FROM alpine:3
WORKDIR /app
COPY --from=base /src/heavenorhell .
EXPOSE 8080
CMD ["./heavenorhell"]