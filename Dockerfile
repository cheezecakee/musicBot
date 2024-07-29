# Stage 1: Build and Install Dependencies
FROM golang:alpine as builder

# Set the Current Working Directory inside the container
WORKDIR /app

ADD . ./

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./main.go


FROM alpine:3.16

ENV DISCORD_BOT=""
ENV SPOTIFY_ID=""
ENV SPOTIFY_SECRET=""
ENV YOUTUBE_API=""
ENV APP_HOME=/app
WORKDIR ${APP_HOME}

# Install required packages
RUN apk --update add --no-cache ca-certificates ffmpeg opus python3

COPY --from=builder /app/main ./main

RUN wget --no-check-certificate https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -O /usr/local/bin/yt-dlp
RUN chmod a+rx /usr/local/bin/yt-dlp
RUN ln -s /usr/bin/python3 /usr/bin/python
RUN yt-dlp --version

# Command to run the executable
ENTRYPOINT ["/app/main"]


