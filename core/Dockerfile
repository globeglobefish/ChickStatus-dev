# Build frontend
FROM node:18-alpine AS frontend-builder
WORKDIR /app/frontend
COPY web/frontend/package*.json ./
RUN npm install
COPY web/frontend/ ./
RUN npm run build

# Build backend
FROM golang:1.21-alpine AS backend-builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o /probe-core ./cmd/core

# Final image
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=backend-builder /probe-core .
COPY --from=frontend-builder /app/dist ./web/dist
EXPOSE 8080
CMD ["./probe-core"]
