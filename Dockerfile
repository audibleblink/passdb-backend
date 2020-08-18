FROM golang:1.15

WORKDIR /app
COPY go.* .
RUN go mod download

COPY . .
CMD ["go", "run", "main.go"]

# $ docker build -t passdb-server .
# $ docker run --env-file .env passdb-server
