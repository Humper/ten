Tradeoffs:

Assume that the # of TOR exit nodes is small enough that we can load them all into memory; making updates easier when we re-download the list. Would have to do something more complex if this weren't true (e.g., double-buffering with versions or something)

No automated testing, in the interest of time. Lots of manual testing.

No canned data / mock tor IP server for local testing; relies on production data for development.

Could be some wonkiness if exploring pagination while doing deletes / updates. Would need versioning + double buffering for this, but hard to get around this absolutely.

Leader election so only one node is doing the updating

Test coverage is somewhat poor -- would probably make integration tests to run against a docker-compose setup to hit the actual postgres database and do some more automated IP exclusion testing.

Country list hard coded in react because I don't know how to dynamically populate it from an API endpoint (react n00b)
