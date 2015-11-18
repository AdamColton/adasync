## Rune Sync
Simple directory synchronization.

To use, set Rune Sync to run as a background task at boot. Rune Sync will periodically scan directories for new collections, out of date collections and collections that need to be synchronized.

-- Todo --
* resolve
** len = shorter
** ordering
** one deletes
** both delete
** move (always go with the longer)
* path collision (different resources want the same location)
* has done sync: if we copy just the ID into an existing dir, it wont have dirTags. need to copy those by location, once we've done that onces, it can be used as a source. If we're the origin, this is set to true

possible issue with dirs: make sure that if I have a dir, rename it and add a new dir in the same location, they don't end up with the same ID

-- Someday --
* hide dot files on windows

### Tools
* https://github.com/go-fsnotify/fsnotify : rough and out of date

### Terms
- Shallow check: compare timestamp, filesize and path
- Deep check: replaces the timestamp check in shallow with a hash compare
- hash: we'll use md5 for now, it may be a bit slow, we'll look into other options later
- Collection ID: every collection is given a crypto-random ID. Only two collections with the same ID can sync.

### Architecture
The root of a runesync dir tree has two files:
- config.collection
- .collection
To create a new collection, a user only needs to create config.collection, they do not need to add anything to it. The .collection file is a binary file not intended for user consumption. This tracks the state and history.

The collection is a collection of resources. A source is a file or directory.
Resource
- ID
- Hash
- Size uint64
- Paths

Path
- Name
- Parent (Resource ID, 0: deleted, 1: root)

Paths is a list of previous paths, a list of moves with the last entry being the current location. If dup=F, a file ID is the hash. If dup=T the file ID will be MD5(hash+initial path). This can still create some odd side effects if multiple collections are uploading and deleteing the same file, starting in the same position. Only a directory is allowed to have size 0, all empty files will be ignored. The hash of a directory will the hash of it's contents hashes, orderd and hashed.

Directories can get tricky. If to users copy the same directory (in terms of bytes) into sperate instances and then sync them, we may end up with two copies. But many naive solutions to this could cause all folders that start as an empty "New Folder" (the default on Windows) to be the same. Rather than solve this now, I'm going to shoot for MVP. Every directory will be given a random ID when it is first added. This may cause duplication, but that shouldn't be too bad.

Runesync has the following procedures that run on different schedules:
- Shallow check known collections for changes
- Deep check known collections for changes (by default, never)
- Scan for collections

Collection checks are done across all instances of a collection. When a collection check comes back with changes we log the additions and deletes. For all the additions, we find the hashes. We then go through the deletes, for each delete we check if we have a matching hash in the additions. If so, we combine them to a move. Otherwise, we execute the delete. When there are no more deletes, we log the additions.

When we find a collection, we start by doing a deep check. When that's done, we get the Collection ID. If we already have an instance of that collection and the last sync node is different, we need to sync them. 

#### PathNode conventions
The two edge cases are if a resource is deleted or is root. In either case Parent is nil. We track what it is by name, either ".root" or ".trash".

### Config Options

#### Instance
* AllowDuplicates: allow more than one copy of a file, regardless of path
* DelayDelete: puts files in a folder
* Ignore: list of expressions
* MaxHistorySize: How long we allow the history to get before we start deleting the oldest nodes
* ArchivesAsBlobs: T = Treat archives (like zips) as blob, otherwise it will inspect the contents
* WindowsFriendly: does not allow two files to only differ by case, does not allow filenames that would break in windows
* AllowDeletes: prevents deletes
* AllowWrites: prevents writes
* SourceOnly: 
* ID: Base64 collection ID. Adding just this is a quick way to sync

#### Service
* Ignore: paths and drives to ignore
* ShallowCheckFreq: how often to do a shallow check
* DeepCheckFreq: how often to do a deep check
* ScanFreq: how often to scan for new drives
* InMemory: [None, Collections, State] state keeps the state of all collections in memory, collections only keeps their locations.

### Future Features
* FTP/SFTP: should be very easy
* Local network: connects to other Runesync servers over udp and keeps shared folders in sync.
* Phone: not sure what the best way to do this is, but I'd like the phones to sync, even better would be a pull-only option for phones
* S3

## Notes
filepath.Walk https://golang.org/pkg/path/filepath/#Walk will be useful for me. It scans all sub directories.