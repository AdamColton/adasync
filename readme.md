## Rune Sync
Simple directory synchronization.

To use, set Rune Sync to run as a background task at boot. Rune Sync will periodically scan directories for new collections, out of date collections and collections that need to be synchronized.

-- Before Check-in --
* config scratch
* clean pub/priv
* mac
* add a version to proto and code
* test remove
** config.collection
** drive
** device

-- Bugs --
* tests are blowing up all over
* restoring a deleted directory does weird things (sync, move sub dir outside collection, sync, move subdir back in)
* Occasionally I'm seening phantom name collisions. Not a big problem because it just duplicates content, but annoying (and concerning)

-- Todo --
* *Sync factory function
* execute actions in depth order
* has done sync: if we copy just the ID into an existing dir, it wont have dirTags. need to copy those by location, once we've done that once, it can be used as a source. If we're the origin, this is set to true
* dirty check for write
* check that dir is still collection before diff
* config
** write formatted config options out

-- Installers --
* win: http://nsis.sourceforge.net/Main_Page

possible issue with dirs: make sure that if I have a dir, rename it and add a new dir in the same location, they don't end up with the same ID

-- Soon --
* file name exclusions
* readOnlyTo
* AllowDuplicates: false

-- Someday --
* hide dot files on windows
* Better recovery logic: small errors can ripple through, it would be good to have a brute force copy fall back
* ftp
* ssh
* local network
* partial hash checks

### Tools
* https://github.com/go-fsnotify/fsnotify : rough and out of date

### Terms
- PathStr: The absolute path as a string to a resource
- Path: The path string broken up into instance path, relative directory and name. The instance path will not end in a slash, the relative directory will start and end in a slash and if the resource is a directory the name will end in a slash
- PathNode: contains a reference to it's parent directory by ID, instance and name. If the parent ID is null and the name is "/", that's the root, if the name ".deleted" that resource has been deleted

### Operation
Collections can do a self sync or a peer sync. They must do a self sync before doing a peer sync.

### Config Options

#### Collection
* AllowDuplicates: allow more than one copy of a file, regardless of path

#### Instance
* Ignore: list of expressions
* MaxHistorySize: How long we allow the history to get before we start deleting the oldest nodes
* ArchivesAsBlobs: T = Treat archives (like zips) as blob, otherwise it will inspect the contents
* WindowsFriendly: does not allow two files to only differ by case, does not allow filenames that would break in windows
* Deletes: full/partial
* readOnly: Does not sync from other drives and does not sync deletes or moves to other drives
* ID: Base64 collection ID. Adding just this is a quick way to sync
* resolveCollision: move/newer;
** move will move one of the files
** newer will try to determine which is more recent and that will become the new instance. This would allow collections of non-media, where the content is changing.

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