FROM golang:1.24-alpine AS builder

WORKDIR /src

COPY main.go .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app main.go

FROM scratch

COPY --from=builder /app /app

EXPOSE 8080

ENTRYPOINT ["/app"]
