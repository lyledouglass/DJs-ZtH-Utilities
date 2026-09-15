FROM golang:latest AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /main .

FROM alpine:latest
RUN apk add --no-cache curl
WORKDIR /
COPY --from=builder /main /main
EXPOSE 8080
CMD [ "/main" ]
HEALTHCHECK --interval=60s --timeout=30s --start-period=30s --retries=3 CMD curl -f http://localhost:8080/healthz || exit 1
