# tasq

Task queue using Redis. Requires Golang 1.18. A Dockerfile and docker-compose
file is provided for reference. It should work out of the box:

```
docker-compose up
```

## Configuration

```
REDIS_URL="127.0.0.1:6379"
REDIS_PASSWORD=""
REDIS_DATABASE="0"
BIND_ADDRESS=":8080"
```

## Deployment

You can also use this `docker-compose.yaml` to run a pre-built `tasq` image,
along with Redis with persistence. This configuration creates a snapshot of your
Redis database every 60 seconds if at least one write happened.

```yaml
version: '3'
services:
  tasq:
    image: ghcr.io/ragtag-archive/tasq:main
    restart: unless-stopped
    environment:
      - REDIS_URL=redis:6379
    ports:
      - 127.0.0.1:8080:8080

  redis:
    image: redis
    restart: unless-stopped
    command: redis-server --save 60 1 --loglevel warning
    volumes:
      - ./data:/data
```

Run with `docker-compose up -d`, and then set up your reverse proxy to point to
`127.0.0.1:8080`.

## API

```
                 _
                | |
 _ __ __ _  __ _| |_ __ _  __ _
| '__/ _' |/ _' | __/ _' |/ _' |
| | | (_| | (_| | || (_| | (_| |
|_|  \__,_|\__, |\__\__,_|\__, |
       _    __/ |          __/ |
      | |  |___/          |___/
      | |_ __ _ ___  __ _
      | __/ _' / __|/ _' |
      | || (_| \__ \ (_| |
       \__\__,_|___/\__, |
                       | |
                       |_|

A basic, easy to use task queue service.
----------------------------------------

PUT /:list
    Add the request body to the list. If the item is already in the list, its
    priority will be bumped up by one.

    Example:
    curl -XPUT -d 'wowzers' https://tasq.url/test
    {
        "ok": true,
        "payload": {
            "key": "test:wowzers"
        },
        "message": ""
    }


GET /:list
    List the first 100 task keys and total count in the specified list, ordered
    by priority from highest to lowest.

    Example:
    curl -XGET https://tasq.url/test
    {
        "ok": true,
        "payload": {
            "tasks": ["test:wowzers"],
            "count": 1
        },
        "message": ""
    }


POST /:list
    Consume an item from the queue. Once consumed, the item will be removed
    from the list. The item with the highest priority will be consumed first.

    Example:
    curl -XPOST https://tasq.url/test
    {
        "ok": true,
        "payload": {
            "key": "test:wowzers",
            "data": "wowzers"
        },
        "message": ""
    }
```
