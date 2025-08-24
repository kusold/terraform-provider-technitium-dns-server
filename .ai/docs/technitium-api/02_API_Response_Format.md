# Technitium DNS Server API - API Response Format

## API Response Format

The HTTP API returns a JSON formatted response for all requests. The JSON object returned contains `status` property which indicate if the request was successful.

The `status` property can have following values:
- `ok`: This indicates that the call was successful.
- `error`: This response tells the call failed and provides additional properties that provide details about the error.
- `invalid-token`: When a session has expired or an invalid token was provided this response is received.

A successful response will look as shown below. Note that there will be other properties in the response which are specific to the request that was made.

```
{
	"status": "ok"
}
```

In case of errors, the response will look as shown below. The `errorMessage` property can be shown in the UI to the user while the other two properties are useful for debugging.

```
{
	"status": "error",
	"errorMessage": "error message",
	"stackTrace": "application stack trace",
	"innerErrorMessage": "inner exception message"
}
```
