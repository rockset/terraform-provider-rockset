# Rockset Terraform Provider Test Resources

This creates all resources that will be used by the test suites of the Rockset Terraform provider.

E.g.

- S3 buckets and files
- Dynamo tables and content
- GCP buckets and files
- Rockset resource fixtures for import tests

# GCS
The test suite assumes you have a base64 encoded string with Google Cloud credentials loaded into env var `TF_VAR_GCS_SERVICE_ACCOUNT_KEY`.

# MongoDB

Free tier MongoDB Atlas clusters cannot be created by the api.

The test suite assumes you have a free tier MongoDB Atlas cluster setup with sample data sets loaded in.

A user with access to this cluster is needed and a connection string using that user credentials must be loaded into env var `TF_VAR_MONGODB_CONNECTION_URI`.

Lock down the cluster using the IPs specified in the console. They're also listed here:
https://docs.rockset.com/mongodb-atlas/#step-4-add-rockset-ips-to-ip-access-list