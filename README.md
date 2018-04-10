## Blockstack Indexers

This is a `golang` implementation of a Profile Indexer for Blockstack. It is meant as a demonstration of a method of building an indexer for Blockstack profiles. Below is a rough list of the steps required to resolve to a profile and the required endpoints to serve the profiles.

1. Fetch all the names on the network
1. Fetch the zonefile for each name
1. Fetch the user's profile from the storage pointed to in their profile
1. Persist that `map[name]profile` in a database
1. Serve the data

I'll be referring to the [`core.blockstack.org`](https://core.blockstack.org/) API in this guide. The one necessary dependency is `blockstack-core`.

### Fetch all names on the network:

First fetch the list of all namespaces using the `/v1/namespaces` endpoint. Then iterate through those namespaces calling the `/v1/namespaces/{tld}/names` route until you have fetched all the names in each namespace. This fetches a full list of all the names on the network. You will then need to persist and update that list. This indexer (as well as core.blockstack.org) does that by writing a `names.json` file that contains a full list of names.

> NOTE: This method doesn't support subdomains. Once the [linked issue](https://github.com/blockstack/blockstack-core/issues/789) is resolved and in master then there will be an additional endpoint that will also need to be polled for names: `/v1/names/sponsored` Those names will need to be incorporated in the `names.json` file.

### Fetch zonefiles for each name:

Next you will need to iterate through those names and pull all the associated zonefiles. The `/v1/name/{name}` will help you with this. Each of these calls will take between 70-200ms and will take some time. One way to reduce this time is to cache the zonefiles on the indexer and run another loop (like the names loop) to update them.

> NOTE: The zonefiles are returned in an RFC compliant format. They can easily be parsed by standard zonefile parsing libraries. This implementation uses the [`miekg/dns`](https://github.com/miekg/dns) library. There are also libraries in pretty much any programming language you would like to write in. [Here's one in Javascript](https://github.com/elgs/dns-zonefile).

### Fetch the user's profile:

Once you have the user's zonefile parsed, you then need to get a list of the `URI` resource records. These records are how blockstack stores references to the user's storage. First fetch each of the URI records from the user's zonefile. You will then [need to decode the `tokenFile`, verify it and and pull out the profile](https://github.com/blockstack/blockstack.js/tree/master/src/profiles). The profile lives in the `claim` section of the decoded `tokenFile`.

> NOTE: The current implementation _does not_ verify profiles. That is a WIP

### Persist `map[name]profile` in a database:

This indexer (as well as the [`blockstack-core` implementation](https://github.com/blockstack/blockstack-core/tree/master/api)) uses mongodb for that. There are details on schema for these in the following languages:

- [Javascript](https://github.com/blockstack/blockstack.js/tree/master/src/profiles/profileSchemas)
- [Golang](/indexer/models.go)

### Serve the data

Write a server that serves this data. You can start by modeling the two `core.blockstack.org` endpoints that rely on this data:

- [`/v1/users/{username}`](https://core.blockstack.org/#resolver-endpoints-lookup-user) - Returns the user's profile
- [`/v1/search?query={query}`](https://core.blockstack.org/#resolver-endpoints-profile-search) - Returns `[]Profile` of names that match the query string

### Work left on this implementation:

- Validate profiles before inserting.
- Serve the data via `core.blockstack.org` compatible API.
