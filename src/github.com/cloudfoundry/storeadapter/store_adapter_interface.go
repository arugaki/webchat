package storeadapter

type StoreAdapter interface {
	// Intiailize connection to server. For a store with no
	// persistent connection, this effectively just tests connectivity.
	Connect() error

	// Create a node and fail if it already exists.
	Create(StoreNode) error

	// Create a dir and fail if it already exists.
	CreateDir(StoreNode) error

	// Update a node and fail if it does not already exist.
	Update(StoreNode) error

	// Set multiple nodes at once. If any of them fail,
	// it will return the first error.
	SetMulti(nodes []StoreNode) error

	// Retrieve a node from the store at the given key.
	// Returns an error if it does not exist.
	Get(key string) (StoreNode, error)

	// Recursively get the contents of a key.
	ListRecursively(key string) (StoreNode, error)

	// Delete a set of keys from the store. If any fail to be
	// deleted or don't actually exist, an error is returned.
	Delete(keys ...string) error

	// Set the ttl on a directory
	UpdateDirTTL(key string, ttl uint64) error

	// Watch a given key recursively for changes. Events will come in on one channel, and watching will stop when a value is sent over the stop channel.
	//
	// Events may be missed, but the watcher will do its best to continue.
	//
	// Returns an error if the watcher cannot initially "attach" to the stream.
	//
	// Otherwise, the caller can assume that the watcher will continue attempting to stream events.
	Watch(key string) (events <-chan WatchEvent, stop chan<- bool, errors <-chan error)

	// Close any live persistent connection, and cleans up any running state.
	Disconnect() error

	// Create a node, keep it there, and send a notification when it is lost. Blocks until the node can be created.
	//
	// To release the node, send a channel value to the releaseNode channel, and read from the channel to ensure it's released.
	//
	// If the store times out, returns an error.
	MaintainNode(storeNode StoreNode) (lostNode <-chan bool, releaseNode chan chan bool, err error)
}
