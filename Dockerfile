ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

FROM alpine:latest

# Copy built files
COPY dist/do-dyndns-${TARGETOS}-${TARGETARCH}${TARGETVARIANT} /usr/bin

# Move built files
RUN mv /usr/bin/do-dyndns-${TARGETOS}-${TARGETARCH}${TARGETVARIANT} /usr/bin/do-dyndns && \
  chmod +x /usr/bin/do-dyndns

# Install Tini
RUN apk --no-cache --no-progress add tini

# Create custom entrypoint supports environment variables
RUN printf "#!/bin/ash\ndo-dyndns" > /entrypoint.sh && \
  chmod +x /entrypoint.sh

ENTRYPOINT ["/sbin/tini", "-g", "--", "/entrypoint.sh"]
