Json API Source
===============

## Configuration

```text
json+api://my.endpoint.dev/v1/sites?token=1234567890
```

MUST return a json with all the sites as:

```json
[
  {
    "id": "one",
    "domain": "one.cloud.localhost",
    "status": "enabled",
    "params": {
      "param.one": "foo",
      "param.two": "bar"
    }
  },
  {
    "id": "two",
    "domain": "two.cloud.localhost",
    "status": "disabled",
    "params": {
      "param.one": "foo",
      "param.two": "bar"
    }
  }
]
```

If the response to this call contains a header like:

```text
Link: <https://my.endpoint.dev/v1/sites/events>; rel="event-stream"
```

This Source will connect to that event stream to obtain changes about the sites.

Events MUST conform to the following structure:

```text
event: add-site
data: {"id":"three","domain":"three.cloud.localhost","status":"enabled","params":{"param.one":"foo","param.two":"bar"}}

event: del-site
data: {"id":"one","domain":"one.cloud.localhost","status":"enabled","params":{"param.one":"foo","param.two":"bar"}}
```

Any events not conforming to the above structure will be silently ignored.