FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o screen-go ./main.go

FROM chromedp/headless-shell:stable AS production
COPY --from=builder /app/screen-go /app/screen-go
EXPOSE 8080
ENTRYPOINT ["/app/screen-go", "serve"]