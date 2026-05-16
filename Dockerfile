# build stage
FROM golang:1.25 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o todo-list-service cmd/todo_list/main.go

# runtime stage
FROM alpine:3.22.4

WORKDIR /root
COPY --from=builder /app/todo-list-service .

EXPOSE 8080
CMD ["./todo-list-service"]
