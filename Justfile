# Set default variables
registry := "ghcr.io"
github_user := env_var_or_default("GITHUB_USER", "dictybase-docker")
ghcr_image := registry + "/" + github_user + "/arangoadmin"
platforms := "linux/amd64,linux/arm64"
dockerfile := "Dockerfile"
tag := "latest"

# Display help
help:
    @just --list

# Build for single architecture (default: linux/amd64)
@build arch="linux/amd64":
    echo "Building for {{arch}}..."
    docker buildx build \
        --platform {{arch}} \
        -f {{dockerfile}} \
        --output type=oci \
        .

# Build for all supported architectures (local only, no push)
build-all:
    #!/usr/bin/env bash
    set -e
    echo "Building for all architectures: {{platforms}}"
    docker buildx build \
        --platform {{platforms}} \
        -f {{dockerfile}} \
        --output type=oci \
        .

# Build and load to local Docker (only works with single architecture)
load arch="linux/amd64":
    #!/usr/bin/env bash
    set -e
    echo "Building and loading for {{arch}}..."
    docker buildx build \
        --platform {{arch}} \
        -f {{dockerfile}} \
        -t {{ghcr_image}}:{{tag}} \
        --load \
        .

# Push single architecture to GHCR
push-ghcr-arch arch="linux/amd64" tag=tag:
    #!/usr/bin/env bash
    set -e
    echo "Authenticating with GHCR..."
    echo $GITHUB_REGISTRY_TOKEN | docker login {{registry}} -u {{github_user}} --password-stdin

    echo "Building and pushing {{arch}} to {{ghcr_image}}:{{tag}}..."
    docker buildx build \
        --platform {{arch}} \
        -f {{dockerfile}} \
        -t {{ghcr_image}}:{{tag}}-{{arch}} \
        --push \
        .

    echo "✓ Successfully pushed {{arch}} image"

# Push all architectures to GHCR with manifest
push-ghcr tag=tag:
    #!/usr/bin/env bash
    set -e
    echo "Authenticating with GHCR..."
    echo $GITHUB_REGISTRY_TOKEN | docker login {{registry}} -u {{github_user}} --password-stdin

    echo "Building and pushing all architectures to {{ghcr_image}}:{{tag}}..."
    docker buildx build \
        --platform {{platforms}} \
        -f {{dockerfile}} \
        -t {{ghcr_image}}:{{tag}} \
        --push \
        .

    echo "✓ Successfully pushed all architectures"
    echo "Image available at: {{ghcr_image}}:{{tag}}"

# Create buildx builder (required for multiarch builds)
setup-builder:
    #!/usr/bin/env bash
    set -e
    builder_name="arangoadmin-builder"

    if docker buildx ls | grep -q "^$builder_name"; then
        echo "Builder '$builder_name' already exists"
    else
        echo "Creating buildx builder: $builder_name"
        docker buildx create --name $builder_name --use
    fi

    echo "Inspecting builder..."
    docker buildx ls

# Inspect current builder
inspect-builder:
    docker buildx ls
    @echo ""
    docker buildx du

# Clean builder cache
clean-cache:
    #!/usr/bin/env bash
    builder=$(docker buildx ls | grep '\*' | awk '{print $1}')
    if [ -n "$builder" ]; then
        echo "Cleaning cache for builder: $builder"
        docker buildx du --builder=$builder || true
        docker buildx prune --builder=$builder --all --force
    else
        echo "No active builder found"
    fi

# Full workflow: setup, build, and push
release tag="latest":
    #!/usr/bin/env bash
    set -e
    echo "=== ArangoAdmin Release: {{tag}} ==="

    # Check environment
    if [ -z "$GITHUB_REGISTRY_TOKEN" ]; then
        echo "Error: GITHUB_REGISTRY_TOKEN not set"
        exit 1
    fi

    echo ""
    echo "1. Setting up builder..."
    just setup-builder

    echo ""
    echo "2. Building and pushing {{tag}}..."
    just push-ghcr {{tag}}

    echo ""
    echo "=== Release complete! ==="
    echo "Image: {{ghcr_image}}:{{tag}}"

# Verify authentication with GHCR
verify-auth:
    #!/usr/bin/env bash
    if [ -z "$GITHUB_REGISTRY_TOKEN" ]; then
        echo "Error: GITHUB_REGISTRY_TOKEN not set"
        exit 1
    fi

    echo "Testing GHCR authentication..."
    echo $GITHUB_REGISTRY_TOKEN | docker login {{registry}} -u {{github_user}} --password-stdin
    echo "✓ Successfully authenticated with GHCR"

# Show current configuration
config:
    @echo "=== ArangoAdmin Docker Build Configuration ==="
    @echo "Registry: {{registry}}"
    @echo "User: {{github_user}}"
    @echo "Image: {{ghcr_image}}"
    @echo "Platforms: {{platforms}}"
    @echo "Dockerfile: {{dockerfile}}"
    @echo "Default Tag: {{tag}}"
    @echo ""
    @echo "Environment:"
    @echo "  GITHUB_USER: {{github_user}}"
    @echo "  GITHUB_REGISTRY_TOKEN: $([ -z \"$GITHUB_REGISTRY_TOKEN\" ] && echo 'NOT SET' || echo 'SET')"
