############################
# STEP 1 build executable binary
############################
# See: https://github.com/chemidy/smallest-secured-golang-docker-image

# golang alpine 1.12
FROM golang@sha256:8cc1c0f534c0fef088f8fe09edc404f6ff4f729745b85deae5510bfd4c157fb2 as builder

# Install make, git + SSL ca certificates.
# Git is required for fetching the dependencies.
# Ca-certificates is required to call HTTPS endpoints.
RUN apk update && apk add --no-cache make git ca-certificates tzdata && update-ca-certificates

# Create appuser
RUN adduser -D -g '' appuser

WORKDIR /app
COPY . .

# Fetch dependencies.
RUN go get -d -v

# Build the binary
RUN cd /app && make build-for-docker

#########################
# STEP 2 build the image
#########################
FROM scratch

# Import from builder.
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd

# Copy our static executable
COPY --from=builder /app/xroad-mock-proxy /xroad-mock-proxy

# Use an unprivileged user.
USER appuser

ENTRYPOINT ["./xroad-mock-proxy"]
