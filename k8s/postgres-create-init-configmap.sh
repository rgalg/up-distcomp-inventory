SQL_FILE="postgres-config/db_schema.sql"

if [ ! -f "$SQL_FILE" ]; then
    echo "Error: SQL schema file not found at $SQL_FILE"
    exit 1
fi

echo "Creating PostgreSQL init ConfigMap from $SQL_FILE..."

# Create ConfigMap directly from the SQL file
kubectl create configmap postgres-init-script \
  --from-file=init.sql="$SQL_FILE" \
  --namespace=inventory-system \
  --dry-run=client -o yaml | kubectl apply -f -

if [ $? -eq 0 ]; then
    echo "ConfigMap postgres-init-script created/updated successfully"
else
    echo "Error: Failed to create ConfigMap"
    exit 1
fi
