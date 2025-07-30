FROM golang:1.23.0-bookworm AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o /myapp .

FROM alpine:latest

COPY --from=build /myapp /myapp
COPY --from=build /app/keys.json /keys.json
COPY --from=build /app/public /public

# Set the MODE environment variable
ENV MODE=prod

ENTRYPOINT ["/myapp"]

# gcloud builds submit --tag gcr.io/maze-conquest-api/maze-conquest-api .