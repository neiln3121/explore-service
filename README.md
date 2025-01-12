# Explore Service

* a gRPC service with four endpoints that
- Lists all users who liked the recipient
- Lists all users who liked the recipient excluding those who have been liked in return
- Counts the number of users who liked the recipient
- Records the decision of the actor to like or pass the recipient 

The service implements cursor based pagaination using an auto increment ID.

The service uses a potgres DB container to store all the data. 

Each decision in the database has a 'liked' and a 'mutually_liked' column. The 'liked' field determines if the actor liked the recipient. The 'mutually_liked' column determines if the recipient liked the actor. If the 'mutually_liked' field is null, then that means that the recipient hasn't given a decision (yet) for the actor. We check if the 'mutually_liked' is null when returing all the users who liked the recipient excluding those that have been liked/not_liked in return.

## Deliverables

To build

`docker-compose build`

To run
`docker-compose up`
The sevice will restart a few times until the DB is running. There are some migrations that automatically run which will create all the required tables and seed some test data. These are located in the database/migrations folder.

To test

`go test ./...`

