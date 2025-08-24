# Technitium DNS Server API - API Request

## API Request

Unless it is explicitly specified, all HTTP API requests can use both `GET` or `POST` methods. When using `POST` method to pass the API parameters as form data, the `Content-Type` header must be set to `application/x-www-form-urlencoded`. When the HTTP API call is used to upload files, the call must use `POST` method and the `Content-Type` header must be set to `multipart/form-data`.

Note! The "set" type of API requests will overwrite any existing value managed by that call. The "add" type of API requests will append to existing value managed by the call.
