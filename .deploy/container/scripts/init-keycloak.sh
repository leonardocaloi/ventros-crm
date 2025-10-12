#!/bin/bash
# Keycloak Initialization Script
# Configura realm, client e roles automaticamente
#
# Este script deve ser executado após o Keycloak estar rodando.
# Usage: ./init-keycloak.sh

set -e

# ====================
# Configuration
# ====================
KEYCLOAK_URL="http://localhost:8180"
ADMIN_USER="admin"
ADMIN_PASSWORD="admin123"
REALM="ventros-crm"
CLIENT_ID="ventros-api"
CLIENT_SECRET="ventros-api-secret"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# ====================
# Helper Functions
# ====================
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Wait for Keycloak to be ready
wait_for_keycloak() {
    log_info "Waiting for Keycloak to be ready..."
    for i in {1..60}; do
        if curl -sf "${KEYCLOAK_URL}/" > /dev/null 2>&1; then
            log_info "Keycloak is ready!"
            sleep 2  # Extra wait for full initialization
            return 0
        fi
        echo -n "."
        sleep 2
    done
    log_error "Keycloak did not become ready in time"
    exit 1
}

# Get admin access token
get_admin_token() {
    log_info "Getting admin access token..."
    TOKEN=$(curl -sf -X POST "${KEYCLOAK_URL}/realms/master/protocol/openid-connect/token" \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "username=${ADMIN_USER}" \
        -d "password=${ADMIN_PASSWORD}" \
        -d "grant_type=password" \
        -d "client_id=admin-cli" | jq -r '.access_token')

    if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
        log_error "Failed to get admin token"
        exit 1
    fi
    log_info "Admin token obtained successfully"
}

# Create Realm
create_realm() {
    log_info "Creating realm: ${REALM}"

    # Check if realm exists
    STATUS=$(curl -sf -o /dev/null -w "%{http_code}" \
        "${KEYCLOAK_URL}/admin/realms/${REALM}" \
        -H "Authorization: Bearer ${TOKEN}")

    if [ "$STATUS" = "200" ]; then
        log_warn "Realm '${REALM}' already exists, skipping creation"
        return 0
    fi

    # Create realm
    curl -sf -X POST "${KEYCLOAK_URL}/admin/realms" \
        -H "Authorization: Bearer ${TOKEN}" \
        -H "Content-Type: application/json" \
        -d '{
            "realm": "'${REALM}'",
            "enabled": true,
            "displayName": "Ventros CRM",
            "displayNameHtml": "<b>Ventros CRM</b>",
            "registrationAllowed": false,
            "resetPasswordAllowed": true,
            "rememberMe": true,
            "verifyEmail": false,
            "loginWithEmailAllowed": true,
            "duplicateEmailsAllowed": false,
            "sslRequired": "none",
            "accessTokenLifespan": 3600,
            "accessTokenLifespanForImplicitFlow": 900,
            "ssoSessionIdleTimeout": 1800,
            "ssoSessionMaxLifespan": 36000,
            "offlineSessionIdleTimeout": 2592000,
            "accessCodeLifespan": 60,
            "accessCodeLifespanUserAction": 300,
            "accessCodeLifespanLogin": 1800,
            "defaultSignatureAlgorithm": "RS256"
        }' > /dev/null

    if [ $? -eq 0 ]; then
        log_info "Realm '${REALM}' created successfully"
    else
        log_error "Failed to create realm '${REALM}'"
        exit 1
    fi
}

# Create Client
create_client() {
    log_info "Creating client: ${CLIENT_ID}"

    # Get client list
    EXISTING_CLIENT=$(curl -sf "${KEYCLOAK_URL}/admin/realms/${REALM}/clients" \
        -H "Authorization: Bearer ${TOKEN}" | jq -r ".[] | select(.clientId==\"${CLIENT_ID}\") | .id")

    if [ -n "$EXISTING_CLIENT" ] && [ "$EXISTING_CLIENT" != "null" ]; then
        log_warn "Client '${CLIENT_ID}' already exists, skipping creation"
        return 0
    fi

    # Create client
    curl -sf -X POST "${KEYCLOAK_URL}/admin/realms/${REALM}/clients" \
        -H "Authorization: Bearer ${TOKEN}" \
        -H "Content-Type: application/json" \
        -d '{
            "clientId": "'${CLIENT_ID}'",
            "name": "Ventros CRM API",
            "description": "Backend API for Ventros CRM",
            "enabled": true,
            "clientAuthenticatorType": "client-secret",
            "secret": "'${CLIENT_SECRET}'",
            "publicClient": false,
            "bearerOnly": false,
            "standardFlowEnabled": true,
            "implicitFlowEnabled": false,
            "directAccessGrantsEnabled": true,
            "serviceAccountsEnabled": true,
            "authorizationServicesEnabled": false,
            "protocol": "openid-connect",
            "attributes": {
                "access.token.lifespan": "3600"
            },
            "redirectUris": [
                "http://localhost:3000/*",
                "http://localhost:8080/*"
            ],
            "webOrigins": [
                "http://localhost:3000",
                "http://localhost:8080"
            ]
        }' > /dev/null

    if [ $? -eq 0 ]; then
        log_info "Client '${CLIENT_ID}' created successfully"
    else
        log_error "Failed to create client '${CLIENT_ID}'"
        exit 1
    fi
}

# Create System Roles
create_roles() {
    log_info "Creating system roles..."

    # Customer role
    log_info "Creating role: customer"
    curl -sf -X POST "${KEYCLOAK_URL}/admin/realms/${REALM}/roles" \
        -H "Authorization: Bearer ${TOKEN}" \
        -H "Content-Type: application/json" \
        -d '{
            "name": "customer",
            "description": "Customer role - Tenant owner with full access to their projects"
        }' > /dev/null 2>&1 || log_warn "Role 'customer' may already exist"

    # Agent role
    log_info "Creating role: agent"
    curl -sf -X POST "${KEYCLOAK_URL}/admin/realms/${REALM}/roles" \
        -H "Authorization: Bearer ${TOKEN}" \
        -H "Content-Type: application/json" \
        -d '{
            "name": "agent",
            "description": "Agent role - User invited to work on projects"
        }' > /dev/null 2>&1 || log_warn "Role 'agent' may already exist"

    log_info "System roles created successfully"
}

# Configure SMTP (optional)
configure_smtp() {
    log_info "Configuring SMTP settings..."

    # For now, we'll skip SMTP configuration
    # User will configure this manually in Keycloak admin console
    log_warn "SMTP configuration skipped - configure manually in Keycloak admin console"
}

# Create test admin user
create_test_user() {
    log_info "Creating test admin user: admin@ventros.com"

    # Check if user exists
    EXISTING_USER=$(curl -sf "${KEYCLOAK_URL}/admin/realms/${REALM}/users?username=admin@ventros.com" \
        -H "Authorization: Bearer ${TOKEN}" | jq -r '.[0].id')

    if [ -n "$EXISTING_USER" ] && [ "$EXISTING_USER" != "null" ]; then
        log_warn "User 'admin@ventros.com' already exists, skipping creation"
        return 0
    fi

    # Create user
    USER_ID=$(curl -sf -X POST "${KEYCLOAK_URL}/admin/realms/${REALM}/users" \
        -H "Authorization: Bearer ${TOKEN}" \
        -H "Content-Type: application/json" \
        -d '{
            "username": "admin@ventros.com",
            "email": "admin@ventros.com",
            "firstName": "Administrator",
            "lastName": "Ventros",
            "enabled": true,
            "emailVerified": true,
            "credentials": [{
                "type": "password",
                "value": "admin123",
                "temporary": false
            }]
        }' -D - | grep -i location | sed 's/.*\///' | tr -d '\r\n')

    if [ -n "$USER_ID" ]; then
        log_info "Test user created successfully"

        # Assign customer role to user
        log_info "Assigning 'customer' role to test user"

        # Get customer role ID
        CUSTOMER_ROLE=$(curl -sf "${KEYCLOAK_URL}/admin/realms/${REALM}/roles/customer" \
            -H "Authorization: Bearer ${TOKEN}" | jq -r '{id: .id, name: .name}')

        # Assign role
        curl -sf -X POST "${KEYCLOAK_URL}/admin/realms/${REALM}/users/${USER_ID}/role-mappings/realm" \
            -H "Authorization: Bearer ${TOKEN}" \
            -H "Content-Type: application/json" \
            -d "[${CUSTOMER_ROLE}]" > /dev/null

        log_info "Role assigned successfully"
    else
        log_warn "Could not retrieve user ID after creation"
    fi
}

# ====================
# Main Execution
# ====================
main() {
    log_info "Starting Keycloak initialization..."
    echo ""

    wait_for_keycloak
    get_admin_token
    create_realm
    create_client
    create_roles
    create_test_user

    echo ""
    log_info "=================================="
    log_info "Keycloak initialized successfully!"
    log_info "=================================="
    echo ""
    log_info "Keycloak Admin Console: ${KEYCLOAK_URL}"
    log_info "Realm: ${REALM}"
    log_info "Client ID: ${CLIENT_ID}"
    log_info "Client Secret: ${CLIENT_SECRET}"
    echo ""
    log_info "Test User:"
    log_info "  Email: admin@ventros.com"
    log_info "  Password: admin123"
    log_info "  Role: customer"
    echo ""
}

# Run main function
main
