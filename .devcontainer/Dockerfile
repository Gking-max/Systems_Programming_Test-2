FROM mcr.microsoft.com/devcontainers/go:1.22-bookworm

# Remove problematic Yarn repository
RUN rm -f /etc/apt/sources.list.d/yarn.list

# Install PostgreSQL client and dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    postgresql-client \
    ca-certificates \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Install migrate CLI tool (version that works with Go 1.22)
RUN go install github.com/golang-migrate/migrate/v4/cmd/migrate@v4.17.0

# Set up Go environment
ENV GOBIN=/usr/local/bin
ENV PATH="$PATH:/usr/local/bin"
ENV GO111MODULE=on

# Create workspace directory
WORKDIR /workspaces/Systems_Programming_Test2