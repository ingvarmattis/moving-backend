#!/bin/bash

# Script for deploying moving-backend to Docker Swarm
# Based on go-environment and godaddy-dyndns deployment patterns

set -e

STACK_NAME="moving-backend"
COMPOSE_FILE="build/app/docker-swarm/docker-swarm.yml"
REGISTRY="us-east4-docker.pkg.dev"
PROJECT_ID="honest-moving"
REPOSITORY="containers"
IMAGE_NAME="backend"

# Get the directory where the script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# Get the backend root directory (parent of scripts directory)
BACKEND_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Change to backend root directory
cd "$BACKEND_ROOT"

echo "🚀 Deploying moving-backend to Docker Swarm..."
echo "📁 Working directory: $(pwd)"

# Verify compose file exists
if [ ! -f "$COMPOSE_FILE" ]; then
    echo "❌ Error: Compose file not found at $COMPOSE_FILE"
    echo "💡 Current directory: $(pwd)"
    echo "💡 Expected path: $(pwd)/$COMPOSE_FILE"
    exit 1
fi

echo "✅ Compose file found: $COMPOSE_FILE"

# Check if .env file exists for environment variables (optional)
if [[ -f .env ]]; then
    echo "📋 Loading environment variables from .env file..."
    export $(cat .env | grep -v '^#' | xargs)
fi

# Check if POSTGRES_URL is set
if [[ -z "$POSTGRES_URL" ]]; then
    echo "❌ Error: POSTGRES_URL must be set!"
    echo "💡 Please set POSTGRES_URL in .env file or export it"
    exit 1
fi

echo "✅ Using provided POSTGRES_URL"

# Check if GCP key file exists
if [ ! -f "gcp-key.json" ]; then
    echo "❌ Error: gcp-key.json file not found!"
    echo "💡 Please provide GCP service account key for registry authentication"
    exit 1
fi

echo "🔍 GCP key file found, validating JSON format..."
if ! python3 -m json.tool gcp-key.json > /dev/null 2>&1; then
    echo "❌ Error: gcp-key.json is not valid JSON!"
    echo "🔍 GCP key file contents preview:"
    head -c 200 gcp-key.json
    exit 1
fi

echo "✅ GCP key JSON format is valid"

# Authenticate to Google Cloud Registry
echo "🔐 Authenticating to Google Cloud Registry..."
echo "🔍 GCP key file size: $(wc -c < gcp-key.json) bytes"
if cat gcp-key.json | docker login -u _json_key --password-stdin $REGISTRY > /dev/null 2>&1; then
    echo "✅ Successfully authenticated to Google Cloud Registry"
else
    echo "❌ Failed to authenticate to Google Cloud Registry"
    echo "🔍 Docker login error:"
    cat gcp-key.json | docker login -u _json_key --password-stdin $REGISTRY 2>&1 || true
    exit 1
fi

# Initialize Docker Swarm if not already initialized
echo "🔧 Checking Docker Swarm status..."
if ! docker info --format '{{.Swarm.LocalNodeState}}' | grep -q "active"; then
    echo "🎯 Initializing Docker Swarm..."
    docker swarm init --advertise-addr $(hostname -I | awk '{print $1}') || echo "⚠️ Swarm may already be initialized"
else
    echo "✅ Docker Swarm is already active"
fi

# Create shared-network if it doesn't exist
echo "🌐 Checking for shared-network..."
if ! docker network ls | grep -q "shared-network"; then
    echo "🔧 Creating shared-network..."
    docker network create --driver overlay --attachable shared-network
else
    echo "✅ shared-network already exists"
fi

# Pull latest image
echo "📥 Pulling latest Docker image..."
docker pull $REGISTRY/$PROJECT_ID/$REPOSITORY/$IMAGE_NAME:latest

# Stop existing stack if running
if docker stack ls | grep -q "$STACK_NAME"; then
    echo "🔄 Stopping existing stack $STACK_NAME..."
    docker stack rm "$STACK_NAME" || true
    echo "⏳ Waiting for services to stop..."
    sleep 15
fi

# Deploy new stack
echo "🚀 Deploying new stack $STACK_NAME..."
docker stack deploy -c "$COMPOSE_FILE" "$STACK_NAME"

# Wait for services to start
echo "⏳ Waiting for services to start..."
sleep 30

# Check stack status
echo "📊 Checking stack status..."
docker stack services "$STACK_NAME"

# Check service health
echo "🔍 Checking service health..."
for service in $(docker stack services "$STACK_NAME" --format "{{.Name}}"); do
    echo "Checking $service..."
    if docker service ls --filter "name=$service" --format "{{.Replicas}}" | grep -q "0/"; then
        echo "⚠️  $service has 0 replicas running"
    else
        echo "✅ $service is running"
    fi
done

# Check application health
echo "🔍 Checking application health..."
sleep 10
if curl -f http://localhost:8001/health > /dev/null 2>&1; then
    echo "✅ Application is responding on port 8001!"
else
    echo "⚠️ Application health check failed, checking logs..."
    docker service logs ${STACK_NAME}_backend --tail 20 2>/dev/null || echo "ℹ️ Logs will be available once service fully starts"
fi

echo ""
echo "🎉 Deployment completed successfully!"
echo "📋 Stack name: $STACK_NAME"
echo "🌐 Service available at:"
echo "   - HTTP: http://localhost:8001"
echo "   - gRPC: localhost:8000"
echo ""
echo "📝 Useful commands:"
echo "   - View logs: docker service logs ${STACK_NAME}_backend"
echo "   - Check status: docker stack services $STACK_NAME"
echo "   - Scale service: docker service scale ${STACK_NAME}_backend=<replicas>"
echo "   - Remove stack: docker stack rm $STACK_NAME"

