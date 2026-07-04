FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /cue ./cmd/cue

FROM alpine:3.21
RUN apk add --no-cache yt-dlp ffmpeg
COPY --from=build /cue /usr/local/bin/cue
WORKDIR /config
ENTRYPOINT ["cue"]