# App Utils

Golang library for micro services in HSM k8s cluster.

It includes support for:

- Server with http/REST routes
- Event driven pub/sub broker with RabbitMQ
- App context with shared correlation id
- Log formatting and context a aware

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
