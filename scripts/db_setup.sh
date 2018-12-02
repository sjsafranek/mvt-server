sudo -u postgres psql -c "CREATE USER geodev WITH PASSWORD 'dev'"
sudo -u postgres psql -c "CREATE DATABASE geodev"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE geodev to geodev"
sudo -u postgres psql -c "ALTER USER geodev WITH SUPERUSER"

PGPASSWORD=dev psql -d geodev -U geodev -f database.sql
