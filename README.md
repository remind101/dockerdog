# DockerDog

DockerDog reports interesting events to DataDog as metrics.

DockerDog requires Docker API 1.22 or higher.

## Why should I use this over the DataDog agent?

The DataDog agent has some problems:

1. It doesn't tag events with the attributes of the event (e.g. no `signal` with the `kill` event).
2. It doesn't generate metric counts from the events, which can be useful in alerting.

## Events

DockerDog generates counters for all container and image events, and tags them with the events attributes:

**Container events**

```
docker.events.container.attach
docker.events.container.commit
docker.events.container.copy
docker.events.container.create
docker.events.container.destroy
docker.events.container.die
docker.events.container.exec_create
docker.events.container.exec_start
docker.events.container.export
docker.events.container.kill
docker.events.container.oom
docker.events.container.pause
docker.events.container.rename
docker.events.container.resize
docker.events.container.restart
docker.events.container.start
docker.events.container.stop
docker.events.container.top
docker.events.container.unpause
docker.events.container.update
```

**Image events**

```
docker.events.image.delete
docker.events.image.import
docker.events.image.pull
docker.events.image.push
docker.events.image.tag
docker.events.image.untag
```
