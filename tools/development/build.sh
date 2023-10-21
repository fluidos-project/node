#!/usr/bin/bash

# Build and load the docker image
build_and_load() {
  local COMPONENT="$1"
  docker build -f ../../build/common/Dockerfile --build-arg COMPONENT="$COMPONENT" -t "$NAMESPACE"/"$COMPONENT":"$VERSION" ../../
  kind load docker-image "$NAMESPACE"/"$COMPONENT":"$VERSION" --name=fluidos-provider
  kind load docker-image "$NAMESPACE"/"$COMPONENT":"$VERSION" --name=fluidos-consumer
}

# Get the Docker namespace, version, and component from the command line
NAMESPACE="$1"
VERSION="$2"
COMPONENT="$3"
VALID_COMPONENTS=("rear-controller" "rear-manager" "local-resource-manager")

# Validate input arguments
if [[ -z "$NAMESPACE" || -z "$VERSION" ]]; then
  echo "Syntax error: ./build.sh <docker_namespace> <version> [<component> ...]"
  exit 1
fi

# Function to check if the component is valid
component_is_valid() {
  local component="$1"
  for valid in "${VALID_COMPONENTS[@]}"; do
    if [[ "$valid" == "$component" ]]; then
      return 0
    fi
  done
  return 1
}

# Build for a specific component or for all components if not specified
if [[ -z "$COMPONENT" ]]; then
  for item in "${VALID_COMPONENTS[@]}"; do
    build_and_load "$item"
  done
elif component_is_valid "$COMPONENT"; then
  build_and_load "$COMPONENT"
else
  echo "Error: Invalid component '$COMPONENT'. Valid components are: ${VALID_COMPONENTS[*]}"
  exit 1
fi
