# load environment variables from .env file
if [ ! -f .env ]; then
    echo "Error: .env file not found!"
    echo "Please create .env file with database credentials."
    exit 1
fi

source .env

OUTPUT_FILE="k8s/postgres-config-generated.yaml"

# generate a YAML file to create a ConfigMap and Secret
cat > "$OUTPUT_FILE" <<EOF
# This file is auto-generated from .env - DO NOT EDIT MANUALLY
# Run k8s/generate-postgres-config.sh to regenerate
apiVersion: v1
kind: ConfigMap
metadata:
  name: postgres-config
  namespace: inventory-system
data:
  DB_HOST: "postgres"
  DB_PORT: "${DB_PORT}"
  DB_NAME: "${DB_NAME}"
  DB_USER: "${DB_USER}"
---
apiVersion: v1
kind: Secret
metadata:
  name: postgres-secret
  namespace: inventory-system
type: Opaque
stringData:
  DB_PASSWORD: "${DB_PASSWORD}"
EOF

echo "PostgreSQL configuration generated: $OUTPUT_FILE"
echo "To apply: kubectl apply -f $OUTPUT_FILE"
