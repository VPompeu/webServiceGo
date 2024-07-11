#!/bin/bash

# Start the Cloud SQL Proxy
/cloud_sql_proxy -dir=/cloudsql &

# Start the Go application
./main
