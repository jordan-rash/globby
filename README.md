# Globby

This actor (we like to call it "Little Globby Tables") is a simple file server showing the basic
CRUD operations of the `wasmcloud:blobstore` contract.

Not only is this actor an example, it is also a fully-functional, HTTP-based fileserver that can be
fronted with any HTTP server implementation and any blobstore implementation (i.e. you could store
the uploaded files on a filesystem or in an s3 compatible store). It also has a full example of
integration testing for the actor

## Required Capability Claims

1. `wasmcloud:httpserver` to receive http requests
2. `wasmcloud:blobstore` to save the image to a blob
3. `wasmcloud:builtin:logging` so the actor can log

## Running this example

This example requires capability providers that fulfill the above contracts. The wasmCloud [HTTP
Server](https://github.com/wasmCloud/capability-providers/tree/main/httpserver-rs) and
[Filesystem](https://github.com/wasmCloud/capability-providers/tree/main/blobstore-fs) capability
providers implement this functionality but you're welcome to use any implementation (like the [S3
Blobstore](https://github.com/wasmCloud/capability-providers/tree/main/blobstore-s3)).

Once you've installed **wash** and ran wasmCloud after following the [installation
guide](https://wasmcloud.dev/overview/installation/), you can run this example actor and the
wasmCloud providers with the following commands:

```
$ wash ctl start actor ghcr.io/jordan-rash/globby:0.2.0
# If you use a locally built actor, replace the actor ID below with your own
$ wash ctl link put MBY3COMRDLQYTX2AUTNB5D2WYAH5TUKNIMELDSQ5BUFZVV7CBUUIKEDR VBBQNNCGUKIXEWLL5HL5XJE57BS3GU5DMDOKZS6ROEWPQFHEDP6NGVZM wasmcloud:blobstore "ROOT=/tmp"
$ wash ctl link put MBY3COMRDLQYTX2AUTNB5D2WYAH5TUKNIMELDSQ5BUFZVV7CBUUIKEDR VAG3QITQQ2ODAOWB5TTQSDJ53XK3SHBEIFNK4AYJ5RKAX2UNSCAPHA5M wasmcloud:httpserver "ADDRESS=0.0.0.0:8080"
$ wash ctl start provider wasmcloud.azurecr.io/blobstore_fs:0.3.1 --skip-wait
$ wash ctl start provider wasmcloud.azurecr.io/httpserver:0.17.0 --skip-wait
```

Once everything is up and running, you can run through all of the operations by following the
annotated commands below:

```console
# Create a file with some content
$ echo 'Hello there!' > myfile.txt

# Upload the file to the fileserver
$ curl -H 'Content-Type: text/plain' -v 'http://127.0.0.1:8080/myfile.txt' --data-binary @myfile.txt
*   Trying 127.0.0.1:8080...
* Connected to 127.0.0.1 (127.0.0.1) port 8080 (#0)
> POST /myfile.txt HTTP/1.1
> Host: 127.0.0.1:8080
> User-Agent: curl/7.85.0
> Accept: */*
> Content-Type: text/plain
> Content-Length: 12
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< content-length: 0
< date: Wed, 18 Jan 2023 23:12:56 GMT
<
* Connection #0 to host 127.0.0.1 left intact

# Get the file back from the server
$ curl -v 'http://127.0.0.1:8080/myfile.txt'
*   Trying 127.0.0.1:8080...
* Connected to 127.0.0.1 (127.0.0.1) port 8080 (#0)
> GET /myfile.txt HTTP/1.1
> Host: 127.0.0.1:8080
> User-Agent: curl/7.85.0
> Accept: */*
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< content-length: 13
< date: Wed, 18 Jan 2023 23:24:24 GMT
<
Hello there!
* Connection #0 to host 127.0.0.1 left intact

# Update the file
$ echo 'General Kenobi!' >> myfile.txt
$ curl -H 'Content-Type: text/plain' -v 'http://127.0.0.1:8080/myfile.txt' --data-binary @myfile.txt
*   Trying 127.0.0.1:8080...
* Connected to 127.0.0.1 (127.0.0.1) port 8080 (#0)
> POST /myfile.txt HTTP/1.1
> Host: 127.0.0.1:8080
> User-Agent: curl/7.85.0
> Accept: */*
> Content-Type: text/plain
> Content-Length: 29
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< content-length: 0
< date: Wed, 18 Jan 2023 23:25:18 GMT
<
* Connection #0 to host 127.0.0.1 left intact

# Get the file again to see your updates
$ curl -v 'http://127.0.0.1:8080/myfile.txt'
*   Trying 127.0.0.1:8080...
* Connected to 127.0.0.1 (127.0.0.1) port 8080 (#0)
> GET /myfile.txt HTTP/1.1
> Host: 127.0.0.1:8080
> User-Agent: curl/7.85.0
> Accept: */*
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< content-length: 29
< date: Wed, 18 Jan 2023 23:26:17 GMT
<
Hello there!
General Kenobi!
* Connection #0 to host 127.0.0.1 left intact

# Delete the file
$ curl -X DELETE -v 'http://127.0.0.1:8080/myfile.txt'
*   Trying 127.0.0.1:8080...
* Connected to 127.0.0.1 (127.0.0.1) port 8080 (#0)
> DELETE /myfile.txt HTTP/1.1
> Host: 127.0.0.1:8080
> User-Agent: curl/7.85.0
> Accept: */*
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< content-length: 0
< date: Wed, 18 Jan 2023 23:33:02 GMT
<
* Connection #0 to host 127.0.0.1 left intact

# (Optional) See that the file doesn't exist anymore
$ curl -v 'http://127.0.0.1:8080/myfile.txt'
*   Trying 127.0.0.1:8080...
* Connected to 127.0.0.1 (127.0.0.1) port 8080 (#0)
> GET /myfile.txt HTTP/1.1
> Host: 127.0.0.1:8080
> User-Agent: curl/7.85.0
> Accept: */*
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 404 Not Found
< content-length: 0
< date: Wed, 18 Jan 2023 23:39:07 GMT
<
* Connection #0 to host 127.0.0.1 left intact
```
