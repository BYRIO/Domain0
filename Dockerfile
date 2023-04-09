# Dockerfile for building a container with the latest version 

# Build the binary Step
FROM golang:1.20
WORKDIR /root/src
COPY . /root/src
RUN go mod download && \
    go build 

# Build the frontend Step
FROM node:18
WORKDIR /root/src/frontend
COPY ./frontend /root/src/frontend
RUN npm install -g pnpm && \
    pnpm install && \
    pnpm run build

# Build the container Step
FROM ubuntu:latest
WORKDIR /opt/
COPY --from=0 /root/src/domain0 /opt
COPY --from=1 /root/src/frontend/dist /opt/static
RUN ln -s /root/config/config.yaml /opt/config.yaml && \
    ln -s /root/config/db.sqlite3 /opt/db.sqlite3 && \
    apt update && apt install -y ca-certificates sqlite3 && \
    rm -rf /var/lib/apt/lists/*

CMD ["/opt/domain0"]


