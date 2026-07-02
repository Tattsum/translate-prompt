# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY backend/ ./backend/

RUN CGO_ENABLED=0 go build -tags noembed -trimpath -ldflags="-s -w" -o /server ./backend/cmd/server

FROM gcr.io/distroless/static-debian12:nonroot

ENV LISTEN_HOST=0.0.0.0
ENV ENV=production

COPY --from=builder /server /server

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/server"]
CMD ["--port", "8080"]
