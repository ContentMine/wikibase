Simple library for talking to Wikibase in golang
=================================================

[![GoDoc](http://godoc.org/github.com/ContentMine/wikibase?status.png)](http://godoc.org/github.com/contentmine/wikibase)

This library has two functions:

* Provides a simple wrapper for common calls the to MediaWiki API for creating and protecting articles along with getting edit tokens.
* Provides a simple pseudo-ORM for working with Items and Properties on wikibase - you build your item as a tagged structure using an embedded header, and then you can sync that up to wikibase to write data there.

Currently this library is work in progress, with a bias on writing to wikibase rather than reading, as that's what has been required on the project this was developed for.


Basics
---------

Currently this library assumes you have valid OAuth tokens for client and consumer. This will be resolved shortly.

For basic API usage there are a series of simple calls in wikibase.go. In general page IDs are used in preference of page titles, for consistency with items and property also referred to by IDs.

If you want to create items and properties, then you can create a structure with a `ItemHeader` embedded entry, which you can store the Item ID in, and then use the `property` annotation on all fields you want to be turned into a property. The value of the property annotation should be the label of the property (not the P number, as that will change most likely between production and test servers, so labels are seen as useful abstractions for naming).

TODO: Add examples here.



License
----------

This library is copyright Content Mine Ltd 2018, and released under the Apache 2.0 License.


Dependencies
-------------

Relies on https://github.com/mrjones/oauth
