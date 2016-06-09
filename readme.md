## AdaSync
AdaSync is designed to make folder syncronization easy.

AdaSync works on "collections". A folder can be an instance of a collection. After AdaSync runs, all instance of a collection that it can access will be syncronized.

AdaSync is not backup software - it performs bi-directional syncing.

### Example Use - Music
Lets say Alice and Bob share a music collection. Alice and Bob each have a computer, a phone and an external storage device (thumbdrive or external hard drive). So there are 6 instances of the collection. And AdaSync will be running in two places (on each of the computers, it does not currently run on phones).

In addition, lets say Alice likes to keep the collection organized, while Bob just dumps new files in the root of the collection. But Bob likes how Alice organizes the music, so when she re-organizes, he wants to replicate that organization.

When ever Alice or Bob plug their phone or external drive into their computer, AdaSync will sync the collections on those devices with the collection on the computer. If Alice takes her phone or hard drive and plugs it into Bobs computer (sneaker-net), her collection will be sync'd with Bobs. As Alice moves and renames files, those will be reflected out to Bob's computer.

### Tag Files
Any directory that is in an AdaSync collection will have a file added to it named ".tag.collection". You can ignore these files, but do not delete them unless you are deleting the entire directory. These tags are how AdaSync tracks folders even if you change their name.

### Instructions
To create a new collection, just add a file name "config.collection" to the root folder of the collection.

You will be able to tell that AdaSync has run on that collection because a new file ".collection" will appear and "config.collection" will have have and id added in the file.

If you create a new folder and copy "config.collection", that folder becomes a copy of the collection, and AdaSync will keep them in sync.

### Static and Non-Static collections
When using AdaSync, you need to decide if a collection is static. Static collections tend to be things like movies, music and pictures where the contents of each file never change. Non-static files tend to be things like documents and spreadsheets where the contents change.

Syncing a lot of non-static data can be very slow. This tends not to be a problem because non-static files are often much smaller (a few MB) where static files, like movies, tend to be larger (10's of GB). If you set a collection of movies to non-static, AdaSync will run slow. But if you set a collection of documents to static, changes to those documents will not be sync'd.

By default, collections are static. To make a collection non static add the line "check file hash: true" to config.collection. Be careful - if one instance is static and another is not, the results can be very unpredictable.

#### Collisions
AdaSync is not meant to be version control software, but it does have the basic ability to handle collisions in non-static collections.

Let's say Alice and Bob share a collection of documents. If Alice and Bob each make a different change to "notes.txt", this is a collision. AdaSync can't know what copy is correct. The only way to fully resolve this is to use source control software like Git that handles versioning, or use a "single-source" like Google Docs.

Here is what will happen when AdaSync detects a collision. Alice will receive a copy of Bob's "notes.txt" renamed to something like "0QBDCH_notes.txt" and Bob will also receive a copy of Alice's changes that follows the same convention. On the next pass, the two originals will be renamed. So Alice and Bob will each have something like "0QBDCH_notes.txt" and "0OWPKK_notes.txt" but there will be no "notes.txt".

One thing to be clear on - if Alice makes changes to "foo.txt" and Bob makes changes to "bar.txt" and they then sync their collections, the updates will be sync'd without collision.

### Configs
#### Read Only
Add "readonly: true"

A Read-Only collection will have it's contents copied into a collection, but will not have data copied into it. Also, deleting from a read-only collection will not replicated the delete. But if a file is copied from a read-only collection into a collection and then deleted from the collection, it will not be copied again.

A good example of where this would be useful is the "pictures" directory on a phone that the camera writes to. Let's say we have a "pictures" folder on a computer and we want to copy all new pictures in when we plug in our phone. 

#### Static
Add "static: false"

This will make a collection non-static so that changes with-in files will be sync'd

#### Check File Length
Add "check file length: true"

Generally, this option should not be used. This creates a non-static collection, but it only checks the file length, not the file contents.