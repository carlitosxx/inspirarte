FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /api ./cmd/api

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata libavif-apps
COPY --from=build /api /api
EXPOSE 8080
CMD ["/api"]
