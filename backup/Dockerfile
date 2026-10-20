FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update \
 && apt-get install -y --no-install-recommends mysql-client rclone ca-certificates curl \
 && rm -rf /var/lib/apt/lists/*

# Scheduler that runs the crontab in the foreground and sends job output to
# the container log. https://github.com/aptible/supercronic
ARG TARGETARCH
ARG SUPERCRONIC_VERSION=0.2.50
RUN case "$TARGETARCH" in \
      amd64) SUM=d2d7c97fbfefc763d5bbcbbb3f8d61f9dd3410ce5d05a737c414994b80b8cdd0 ;; \
      arm64) SUM=a1c0af9872e8de89a1ce5c1b596e26c42917d5a2cf914d98faee2e28156899e2 ;; \
      *) echo "unsupported architecture: $TARGETARCH" >&2; exit 1 ;; \
    esac \
 && curl -fsSL -o /usr/local/bin/supercronic \
      "https://github.com/aptible/supercronic/releases/download/v${SUPERCRONIC_VERSION}/supercronic-linux-${TARGETARCH}" \
 && echo "${SUM}  /usr/local/bin/supercronic" | sha256sum -c - \
 && chmod 755 /usr/local/bin/supercronic

COPY backup.sh /usr/local/bin/backup.sh
COPY crontab /etc/backup.crontab
RUN chmod 755 /usr/local/bin/backup.sh

ENV TZ=UTC
ENTRYPOINT ["supercronic", "/etc/backup.crontab"]
