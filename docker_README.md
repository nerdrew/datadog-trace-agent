# Generic setup

Use the following Dockerfile to get up and running with the trace-agent.
This image will soon be released publically

### Dockerfile_apm
```
FROM datadog/docker-dd-agent

MAINTAINER Datadog <package@datadoghq.com>

RUN echo "deb http://apt-trace.datad0g.com.s3.amazonaws.com/ stable main" > /etc/apt/sources.list.d/datadog-trace.list \
 && apt-get update \
 && apt-get install -y ca-certificates \
 && apt-get install --no-install-recommends -y dd-trace-agent \
 && apt-get clean \
 && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

EXPOSE 7777/tcp

ENTRYPOINT ["/opt/datadog-agent/bin/trace-agent"]
```

## Build the image
```
docker build -f Dockerfile_apm . -t dd_trace_agent
```

## Running the container
Set your api key and bind the trace agent to the default route via environment variables

```
docker run --name dd_trace_agent -d -p 7777:7777 -e DD_API_KEY=my_api_key -e DD_BIND_HOST=0.0.0.0 dd_trace_agent
```

The agent accepts a few other configuration values from the environment. See General Configuration for more details

## Reporting to the agent container from the host
Existing clients are configured to report to `localhost:7777` by default, so with the above setup no additional configuration
is required to report to the container.
After installing one of our integrations `docker logs -f dd_trace_agent` should show traces being registered by the agent

## Reporting to the agent on the host from an application within a container
One can either:
- Run your container with `--net=host`. This has security caveats as described in the [Docker documentation](https://docs.docker.com/engine/reference/run/#/network-settings)
- Bind the agent to the host's address on the `docker0` bridge. `sudo ip addr show docker0` should give you this.
E.g.
```
$ sudo ip addr show docker0
4: docker0: <NO-CARRIER,BROADCAST,MULTICAST,UP> mtu 1500 qdisc noqueue state DOWN
    link/ether 02:42:d5:18:f5:5c brd ff:ff:ff:ff:ff:ff
    inet 172.17.0.1/16 scope global docker0
```

In `/etc/dd-agent/datadog.conf`
```
bind_host: 172.17.0.1
```

Or via the environment
```
DD_BIND_HOST=172.17.0.1
```

Next, configure your application tracer to report traces to the same address.
This assumes that the Docker host's own IP address on `docker0` is configured as the default gateway for running containers
(this is the case in vanilla setups. if you've tweaked this, you know what you're doing :))

An example in python;
```
from ddtrace import tracer; tracer.configure(hostname="172.17.0.1")
```

## Reporting to the agent container from another container
Start the agent container with

```
docker run --name dd_trace_agent -d -p 7777:7777 -e DD_API_KEY=my_api_key -e DD_BIND_HOST=0.0.0.0 dd_trace_agent
```

And then configure your application tracer to point to the Docker host's own IP address on `docker0` - just as mentioned in the previous section
```
from ddtrace import tracer; tracer.configure(hostname="172.17.0.1")
```
