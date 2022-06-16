#!/bin/bash
# initialise database for ox-app service
# auto-generated by Artisan on 2022-06-01 15:52:30.100311 +0000 UTC

# configure 'onix' database release information
dbman config use -n onix-config
dbman config set repo.uri https://raw.githubusercontent.com/gatblau/ox-db/d0b7520ce991c74b52efa2c381f001335d343b4a
dbman config set db.provider _pgsql
dbman config set db.host db
dbman config set db.port 5432
dbman config set db.name onix
dbman config set db.username ${OX_APP_DB_USER}
dbman config set db.password ${OX_APP_DB_PWD}
dbman config set db.adminusername postgres
dbman config set db.adminpassword ${DB_POSTGRES_PASSWORD}
dbman config set appversion 0.0.4

# create 'onix' database
dbman db create

# deploy 'onix' database schema
dbman db deploy
