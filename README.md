# core
reusable components across different microservices in serendipity. I attempted to make them as
generic as possible so that they may be of use to others.


### request module
Inspired by the [imroc/req](https://github.com/imroc/req) library but I did not like how they supported testing.
Also, I was only making use of a very small subset of the features so I decided to create a proprietary lightweight version
as well as define a `Mock` client to be used for unit testing.

### logger module

### MongoDB client

### Strava client

### AWS SQS client