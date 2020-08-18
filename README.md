# App Utils

Golang library for micro services in HSM k8s cluster.

It includes support for:

- server: Server with http/REST routes
- eventPubSub: Event driven pub/sub broker with RabbitMQ
- appcontext: App context with shared correlation id
- applog: Log formatting and context a aware
- appSaga: Saga manager for handling sequence of events
- Notification Manager: Manager for push notification
- db: database impl. for postgres and mock

### Running tests

Running the applications on docker compose are needed to execute tests

```
> cd docker
> docker-compose up -d

```
### Docker Compose

In docker folder there is a collection of applications needed
for the tech stack:

- RabbitMQ container
    - manager on http://localhost:15672
    - user: rabbitmq
    - pw: rabbitmq
    - ampq on ampq://rabbitmq:rabbitmq@localhost

- Postgres DB container
    - on localhost:5432
    - user: postgres
    - pw: postgresx
    - db: postgres
    
- Redis container
    - on localhost:6379
       
