# Elastic Data Model
## Users
The index titled users will hold information about users. Only the webservice should have read and write access to this index. The "_id" will be the google provided openid. This will have to change if other oauth providers are added, probably to a concatenation of oauth provider name and provided id. 
```
PUT /users
{
    "mappings": {
        "properties": {
            "hashed_api_key": {"type": "text" }
        }
    }
}
```
## Progress
The index titled progress will hold the progress of timers. It is for internal correctness, so only the elastic consumer writes to it, and only runtimers reads from it. The "_id" will be a stringified protobuf messaging.Key.
```
PUT /progress
{
    "mappings": {
        "properties": {
            "last_execution": { "type": "date" },
            "completed_executions" : { "type" : "integer" }
        }
    }
}
```
## Executions-{domain}
This is a record of all the timer executions for a given user. Written by elastic consumer and only read by the webservice when a user is authenticated for that domain. The "_id" is the UUID of the timer. 
```
PUT /executions-{domain}
{
    "mappings": {
        "properties": {
            "timer_uuid" : {"type" : "text" }
            "data": { "type" : "text" }
        }
    }
}
```
## Timers-{domain}
This is a record of all the timers for a given user. Written by elastic consumer and only read by the webservice when a user is authenticated for that domain. The "_id" will be the uuid of the timer.
```
PUT /timers-{domain}
{
    "mappings": {
        "properties": {
            "task": {
                // we'll just resolve this with dynamic mapping
            },
            "schedule": {
                "properties": {
                    "cron" : { "type" : "text" },
                    "start_time" : { "type" : "date" },
                    "stop_time" : { "type" : "date" },
                    "max_executions" : { "type" : "integer" }
                }
            }
        }
    }
}
```