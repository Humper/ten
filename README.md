# TOR Exit Node Browser

This system is composed of two pieces:

1. A Go backend that fetches lists of TOR Exit Node IP addresses from multiple source of truth and merges them into a postgres database.  It periodically scans this database for any entries without associated geolocation data and fetches those data as well.
2. A react-admin based frontend that supports browsing the data, as well as filtering the data based on country code, and excluding certain IP addresses on a per-user basis.

The backend also manages users with passwords, and associates the list of disallowed IP addresses with those users.

## Design decisions and tradeoffs

* During the exit node update process (which runs hourly), we assume that the number of TOR exit nodes is small enough that we can load them all into memory.  This makes updates much easier when we re-download the list. If this assumption were violated, we would have to do something more complex like double-buffering the list of exit nodes and versioning them.

* Unit test coverage is somewhat spotty; in the interest of time, I haven't written any integration tests and instead done lots of manual testing.

* Because the list of exit nodes updates on a regular cadence, the pagination could be a little weird if updates to the list happen while a user is browsing.  

* I used `etcd` to perform leader election so only one server node is doing the updating.  Of course, the docker-compose setup only has a single server node, so this hasn't really been stress tested.

* This is my first ever react app, so I couldn't figure out how to dynamically populate the list of country codes for filtering; the list is just hardcoded in the javascript.

* The list of IP addresses to omit per user are called "allowed IPs" because the original task statement referred to these as an "allowlist"; this is a little confusing in a couple of places.

## How to test

1. `cd testing; docker-compose up --build` will rebuild and start the backend server.
2. `cd static/ten; npm start` will start the frontend server.

When the database is empty, the server will create a user with email `admin@admin.com` and password `password`.  This user can be used as normal, or to create additional users as desired.
