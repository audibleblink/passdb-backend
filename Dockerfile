# #================
# # Stage 1: Frontend build (Node.js)
# #================
# FROM node:18-alpine AS frontend-builder
# WORKDIR /app
# RUN apk add --no-cache make
# COPY Makefile package*.json ./
# RUN make docs/index.html

#================
# Stage 2: Go build
#================
FROM golang:1.24-alpine AS go-builder
WORKDIR /app
# RUN apk add --no-cache git make
COPY . .
# COPY --from=frontend-builder /app/docs /app
RUN go build -o build/passdb

#================
# Stage 3: Final runtime image
#================
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata curl
WORKDIR /app
COPY --from=go-builder /app/build/passdb /usr/local/bin/passdb
EXPOSE 3000
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:3000/api/v1/health || exit 1

ENTRYPOINT ["passdb"]
