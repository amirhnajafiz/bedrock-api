# Bedrock API

Bedrock API is an HTTP server service that enables interaction with **Bedrock** tools via HTTP requests. This service is responsible for tracing management and log collection.

![Abstract Flow](images/abstract_flow_diagram.svg)

## Modesl

* Request
  * Docker Image
  * Command
  * Timeout
  * Status
    * UID
    * Pending | Running | Stopped | Finished | Failed
    * Uptime
    * Trace bytes

## API Endpoints

* [POST] /api/new
  * Accept a request from user, set the status (uid, pending, uptime), store it in KV storage.
  * The docker daemon gets pending requests and starts the container (the bedrock tracer container first and the target container).
  * The file manager daemon creates the output directory to store the tracing logs and metadata.
    * data/container-name/...
  * The docker daemon updates the status of a request.
  * Once a container is stopped, finished or failed, the docker daemon must do the cleanup process.
* [POST] /api/stop
  * If a request is not finished or failed, the user can stop it.
  * The docker daemon must terminate all related containers.
* [GET] /api
  * The default route of API must return a list of requests with it's current status.
* [GET] /api/id
  * Upon calling the default route with a request UID, we must serve the output files of tracing from the file manager daemon.

## Related Projects

* [Bedrock Tracer](https://github.com/amirhnajafiz/bedrock-tracer)
