FROM golang:tip-alpine3.22 AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o /myapp .

FROM alpine:latest

COPY --from=build /myapp /myapp

# Set the MODE environment variable
ENV MODE=prod

ENTRYPOINT ["/myapp"]

# gcloud builds submit --tag gcr.io/<project-name>/<project-name> .