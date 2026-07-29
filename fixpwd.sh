#!/bin/bash
PW=$(grep -E '^POSTGRES_PASSWORD' ~/deploy/supabase/docker/.env | cut -d= -f2)
psql "postgres://postgres:$PW@127.0.0.1:54322/postgres?sslmode=disable" -f /tmp/update_admin.sql
