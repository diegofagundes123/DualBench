# Ambiente de desenvolvimento / build do DualBench (Wails v2) — Debian Bookworm
FROM golang:1.22-bookworm

RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    libgtk-3-dev \
    libwebkit2gtk-4.1-dev \
    pkg-config \
    curl \
    ca-certificates \
    git \
    && rm -rf /var/lib/apt/lists/*

RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y --no-install-recommends nodejs \
    && rm -rf /var/lib/apt/lists/*

ENV CGO_ENABLED=1
ENV GOFLAGS=-buildvcs=false
RUN go install github.com/wailsapp/wails/v2/cmd/wails@v2.9.2

COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

WORKDIR /app
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
CMD ["wails", "dev"]
