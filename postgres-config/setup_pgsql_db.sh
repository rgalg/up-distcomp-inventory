# load environment
if [ -f .env ]; then
    echo "Loading environment variables from .env file..."
    export $(grep -v '^#' .env | xargs)
else
    echo "Error: .env file not found!"
    echo "Please copy .env.template to .env and configure your credentials."
    exit 1
fi

# check if schema file exists
if [ ! -f postgres-config/db_schema.sql ]; then
    echo "Error: postgres-config/db_schema.sql not found!"
    echo "Please create the schema file first."
    exit 1
fi

echo "Setting up PostgreSQL database..."

# using postgres superuser
sudo -u postgres psql << EOF
-- drop existing database and user if they exist
DROP DATABASE IF EXISTS $DB_NAME;
DROP USER IF EXISTS $DB_USER;

-- create user
CREATE USER $DB_USER WITH ENCRYPTED PASSWORD '$DB_PASSWORD';
-- create database
CREATE DATABASE $DB_NAME OWNER $DB_USER;

-- connect to the database
\c $DB_NAME

-- grant schema permissions
GRANT ALL ON SCHEMA public TO $DB_USER;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO $DB_USER;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO $DB_USER;

-- set default privileges
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO $DB_USER;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO $DB_USER;

EOF

# check if previous command was successful (exit code 0)
if [ $? -ne 0 ]; then
    echo "ERROR: Failed to create database and user"
    exit 1
fi

echo "Database and user created. Applying schema..."

# run the schema SQL file as the new user
PGPASSWORD="$DB_PASSWORD" psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f postgres-config/db_schema.sql

if [ $? -ne 0 ]; then
    echo "ERROR: Failed to apply schema"
    exit 1
fi

echo "Schema applied successfully!"

# verify everything worked
echo ""
echo "Verifying database setup..."
echo "================================"

echo "Tables created:"
PGPASSWORD="$DB_PASSWORD" psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "\dt"

echo ""
echo "Product count:"
PGPASSWORD="$DB_PASSWORD" psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT COUNT(*) as product_count FROM products;"

echo ""
echo "Inventory count:"
PGPASSWORD="$DB_PASSWORD" psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT COUNT(*) as inventory_count FROM inventory;"

echo ""
echo "Sample data:"
PGPASSWORD="$DB_PASSWORD" psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT p.id, p.name, p.price, i.stock, i.reserved FROM products p JOIN inventory i ON p.id = i.product_id;"

echo ""
echo "================================"
echo "====   DB setup completed   ===="
echo "====    I hope it worked    ===="
echo "================================"
