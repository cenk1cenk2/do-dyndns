ARG IMAGE_NAME

FROM ${IMAGE_NAME}

ARG BINARY_NAME

# Copy built files
COPY dist/do-dyndns-${BINARY_NAME} /usr/bin

# Move built files
RUN mv /usr/bin/do-dyndns-${BINARY_NAME} /usr/bin/do-dyndns && \
  chmod +x /usr/bin/do-dyndns

# Install Tini
RUN apk --no-cache --no-progress add tini

# Create custom entrypoint supports environment variables
RUN printf "#!/bin/ash\ndo-dyndns" > /entrypoint.sh && \
  chmod +x /entrypoint.sh

ENTRYPOINT ["/sbin/tini", "-vg", "--", "/entrypoint.sh"]
