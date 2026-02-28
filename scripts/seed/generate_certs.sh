#!/bin/bash
# =============================================================================
# CERTIFICATE GENERATION SCRIPT
# Generates self-signed TLS certificates for development/testing
# =============================================================================

set -e

# Configuration
CERT_DIR="${1:-./certs}"
CERT_FILE="${CERT_DIR}/cert.pem"
KEY_FILE="${CERT_DIR}/key.pem"
DAYS_VALID="${2:-365}"
COMMON_NAME="${3:-localhost}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}TLS Certificate Generation${NC}"
echo -e "${YELLOW}========================================${NC}"

# Create certificate directory
mkdir -p "$CERT_DIR"

# Check if openssl is available
if ! command -v openssl &> /dev/null; then
    echo -e "${RED}Error: openssl is not installed${NC}"
    echo "Please install openssl and try again."
    exit 1
fi

# Generate private key and self-signed certificate
echo -e "${YELLOW}Generating private key and certificate...${NC}"
echo "  Certificate: $CERT_FILE"
echo "  Private Key: $KEY_FILE"
echo "  Valid for: $DAYS_VALID days"
echo "  Common Name: $COMMON_NAME"
echo ""

# Generate certificate with Subject Alternative Name (SAN) for localhost.
# Prefer -addext (newer OpenSSL), fallback to a temp config for older builds.
if ! openssl req -x509 -newkey rsa:4096 \
    -keyout "$KEY_FILE" \
    -out "$CERT_FILE" \
    -days "$DAYS_VALID" \
    -nodes \
    -subj "/C=US/ST=State/L=City/O=Forum/OU=Development/CN=${COMMON_NAME}" \
    -addext "subjectAltName=DNS:localhost,DNS:${COMMON_NAME},IP:127.0.0.1"; then
    echo -e "${YELLOW}OpenSSL -addext unsupported, retrying with config file...${NC}"
    TMP_CONF="$(mktemp)"
    cat > "$TMP_CONF" <<EOF
[req]
distinguished_name = dn
x509_extensions = v3_req
prompt = no

[dn]
C = US
ST = State
L = City
O = Forum
OU = Development
CN = ${COMMON_NAME}

[v3_req]
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
DNS.2 = ${COMMON_NAME}
IP.1 = 127.0.0.1
EOF

    openssl req -x509 -newkey rsa:4096 \
        -keyout "$KEY_FILE" \
        -out "$CERT_FILE" \
        -days "$DAYS_VALID" \
        -nodes \
        -config "$TMP_CONF" \
        -extensions v3_req

    rm -f "$TMP_CONF"
fi

# Set appropriate permissions
chmod 600 "$KEY_FILE"
chmod 644 "$CERT_FILE"

echo -e "${GREEN}✓ Certificates generated successfully!${NC}"
echo ""
echo "Files created:"
echo "  - $CERT_FILE (certificate)"
echo "  - $KEY_FILE (private key)"
echo ""
echo "To use these certificates, update your .env file:"
echo "  TLS_CERT_FILE=${CERT_FILE}"
echo "  TLS_KEY_FILE=${KEY_FILE}"
echo ""
echo -e "${YELLOW}Note: These are self-signed certificates for development.${NC}"
echo -e "${YELLOW}For production, use certificates from a trusted CA.${NC}"
