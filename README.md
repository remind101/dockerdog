# DockerDog

DockerDog reports Docker events to DataDog as metrics.

DockerDog requires Docker API 1.22 or higher.

## Why should I use this over the DataDog agent?

The DataDog agent has some problems:

1. It doesn't tag events with the attributes of the event (e.g. no `signal` with the `kill` event).
2. It doesn't generate metric counts from the events, which can be useful in alerting.

## Events

DockerDog generates these counter metrics for the corresponding Docker events:


| Metric Name                           | Tags       |
| ---                                   | ---        |
| `docker.events.container.attach`      |            |
| `docker.events.container.create`      |            |
| `docker.events.container.destroy`     |            |
| `docker.events.container.detach`      |            |
| `docker.events.container.die`         | `exitCode` |
| `docker.events.container.exec_create` |            |
| `docker.events.container.exec_detach` |            |
| `docker.events.container.exec_start`  |            |
| `docker.events.container.kill`        | `signal`   |
| `docker.events.container.oom`         |            |
| `docker.events.container.start`       |            |
| `docker.events.container.stop`        |            |
| `docker.events.image.delete`          |            |
| `docker.events.image.import`          |            |
| `docker.events.image.load`            |            |
| `docker.events.image.pull`            |            |
| `docker.events.image.push`            |            |
| `docker.events.image.save`            |            |


## Tags from Event Attributes

DockerDog can be configured to map event attributes to metric tags. For
example, the following command includes the `image` attribute as a tag with key
`image`, and the `com.example.tags.app.name` attribute as a tag with key
`app_name`:

```
dockerdog -a image -a com.example.tags.app.name:app_name
```
