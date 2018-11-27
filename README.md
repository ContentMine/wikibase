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

If you want to create items and properties, then you can create a structure with a `ItemHeader` embedded entry, which you can store the Item ID in, and then use the `property` annotation on all fields you want to be turned into a property (you can add additional fields to the structure, if they don't have a property tag then they will be safely ignored). The value of the property annotation should be the label of the property (not the P number, as that will change most likely between production and test servers, so labels are seen as useful abstractions for naming).

```
type ExampleWikibaseItem struct {
	wikibase.ItemHeader

    Name             string                 `property:"Name"`
    Birthday         time.Time              `property:"Date of birth"`
    NextOfKin        *wikibase.ItemProprety `property:"Next of kin,omitoncreate"`
    SkateboardsOwned int                    `property:"Skateboards owner"`
}
```

The library will manage some of the property formatting restrictions of Wikibase: Pointers with a nil value will be set as having `no value` in Wikibase, as will string properties with a zero length. Strings will automatically have whitespace formatting homogenised to keep Wikibase happy too.

The `omitoncreate` modified on the tag will tell the library not to attempt to set an initial value for that property when the item is being created. If you are uploading a set of items and then layer need to link them using ItemProperty fields then you may not wish to load them initially at create time and upload them later as a restricted subset (using the argument to the update call to say only add new items). Ideally this sort of thing wouldn't be necessary but the Wikibase API is relatively slow with even trivial amounts of data, so this lets you start to manage how much you actually do in each transaction.


Once you've defined your structure and created a client you first need to get the Client to loop up the actual P numbers of the properties using a call to `MapPropertyAndItemConfiguration` like so:

```
    client := wikibase.NewClient(...)
    err := client.MapPropertyAndItemConfiguration(ExampleWikibaseItem{}, true)
```

The boolean second argument tells the client to create property definitions if they don't already exist on the wikibase server.

If you want to fetch the Q numbers for specific items so you can store them in `ItemProperty` fields then you can call `MapItemConfigurationByLabel`, which also takes a second argument to say whether it should create the item if not found.

You can create a new Wikibase item as follows:

```
    person := ExampleWikibaseItem{Name: "Alice", Birthday: time.Now()}
    err := client.CreateItemInstance("person item", &person)
```

After this call, assuming successful, the person.ID field will be set to the Q number for the Item on Wikibase created, and the person.PropertyIDs field will map the P numbers of the properties to their GUID on wikibase. It is recommend you serialise these to JSON or some other format and restore them later if you wish to edit the same object on Wikibase across multiple invocations of your client.

You can similarly update the Wikibase Item you are modelling like so:

```
    person.SkateboardsOwned += 1
    err := client.UploadClaimsForItem(&person, true)
```

The boolean argument indicates if properties already uploaded should be updated or ignored. True here means updated, false would have been effectively a no-op. The API is like this as Wikibase API updates are relatively slow, and so having the fidelity to control how much up update can make for a much quicker client.


License
----------

This library is copyright Content Mine Ltd 2018, and released under the Apache 2.0 License.


Dependencies
-------------

Relies on https://github.com/mrjones/oauth
