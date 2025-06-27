FROM golang:1.24.4

WORKDIR /app

# airのバイナリをインストール
RUN go install github.com/air-verse/air@latest

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

CMD ["air"]
