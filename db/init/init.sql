-- kubectl port-forward svc/mypostgres 5436:5432 -n postgres
-- make init
-- user creation
CREATE USER pgq_app WITH PASSWORD 'DB_PASSWORD';

CREATE DATABASE pgq_local;

GRANT ALL PRIVILEGES ON DATABASE pgq_local TO pgq_app;