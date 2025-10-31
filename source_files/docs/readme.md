Project Charlie

This project contains the backend services for the Charlie processing pipeline.

Services

charlie-service-processor: Main data processing unit.

S3Connector: Utility for connecting to our S3 buckets.

Setup

Run ./build.sh to compile the Java code and build the Docker image.

Use the service.yaml to deploy to Kubernetes.