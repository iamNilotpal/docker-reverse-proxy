FROM golang:1.23.1-bullseye AS build

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download && go mod verify

COPY . /app

RUN CGO_ENABLED=0 GOARCH=arm64 GOOS=linux go build -o app -a -ldflags="-s -w" -installsuffix cgo main.go

RUN apt-get update && apt-get install -y --no-install-recommends \
  upx \
  build-essential \
  ca-certificates \
  && rm -rf /var/lib/apt/lists/*

RUN upx --ultra-brute -qq app && upx -t app

FROM scratch

WORKDIR /app

COPY --from=build /app/app .

ENTRYPOINT [ "./app" ]