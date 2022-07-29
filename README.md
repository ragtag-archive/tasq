# tasq

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
    Add the request body to the list.

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
    List available task keys and total count in the
    specified list.

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
    Consume an item from the queue. Once consumed, the item
    will be removed from the list.

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
