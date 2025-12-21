FROM golang:latest AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /main .

FROM alpine:latest
WORKDIR /
COPY --from=builder /main /main
CMD [ "/main" ]
HEALTHCHECK --interval=60s --timeout=30s --start-period=5s --retries=3 CMD curl -f http://discord.gg || exit 1
