## Rune Sync
This is another test.

Simple directory synchronization.

To use, set Rune Sync to run as a background task at boot. Rune Sync will periodically scan directories for new collections, out of date collections and collections that need to be synchronized.

-- Before Check-in --
* clean pub/priv
* test remove
** config.collection
** drive
** device

-- Bugs --
Add folder with content, sync. Rename folder in A. Folder in B is deleted on first pass, syncs on second. It looks like the rename is not registering as a move.

Also, all children are being registered as missing, when we detect a dir move, we need to update that the children moved.

It might be better to do a tag check when we encounter a directory...

-- Todo --
* when checking a file, we could potentially compute the hash twice, that seems wasteful
* Let the user set the collection ID or read only ID
* If config.collection is deleted, delete .collection
* Link to github on download page
* cron to cp from this drive to projects and cron to wget from projects
* *Sync factory function
* has done sync: if we copy just the ID into an existing dir, it wont have dirTags. need to copy those by location, once we've done that once, it can be used as a source. If we're the origin, this is set to true
* check that dir is still collection before diff
* config
** write formatted config options out
* possible issue with dirs: make sure that if I have a dir, rename it and add a new dir in the same location, they don't end up with the same ID

-- Installers --
* win: http://nsis.sourceforge.net/Main_Page

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
* tag directory: false

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
-- Current --
* readOnly
* check file length
* check file hash
* AllowDuplicates (sort of)

-- Future --
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

### Order of operations

-- Self --
# Walk root dir
# Resolve moved dir

-- Sync --
# Create Directories
# Move directories
# Copy files
# Move files
# Delete Directories
# Delete Files
# Attempt Renames

BAHHHH, need to move directories before creating directories in case A->B, new A. but need to delete directories before moving directories incase delete B, A->B but need to create directories before deleting directories incase we're deleting B, but need to copy a resource out of it first. Together, these 3 create a conflict.

NO! We break that cycle with the create collision. If we try to create A, but there's already an A, we create _XXXXX_A and add an 'attemptRename' action.
