FROM golang:alpine AS build

COPY . .

WORKDIR cmd/api

RUN go build -o go-fitness

FROM alpine:3.18

RUN apk update && apk add --no-cache \
    bash \
    ffmpeg \
    ffmpeg-dev

# Set the working directory inside the container
WORKDIR /app


# Copy only the necessary files
COPY  --from=build /go/cmd/api/go-fitness .

# Copy additional configs
COPY  --from=build /go/.env.example .env

# If it is a config also copy it:
COPY  --from=build /go/config/dev.yaml ./config/dev.yaml

COPY  --from=build /go/client_secret.json ./client_secret.json

# If it is a lang also copy it:
COPY  --from=build /go/lang/active.ro.json ./lang/active.ro.json

COPY --from=build /go/docker/dev/entrypoint.sh /sbin/entrypoint.sh
RUN chmod +x /sbin/entrypoint.sh

EXPOSE 8082

LABEL "org.opencontainers.image.source"="https://github.com/dmitriy-robu/ffm"

# Run the application
ENTRYPOINT ["bash", "/sbin/entrypoint.sh"]

CMD ["true"]
