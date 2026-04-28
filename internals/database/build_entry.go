package internals


/*
	When building an entry from our index, we will need to get the pathname and
	the parent directories for this pathname

	So the Tree and the IndexEntry will implement this interface
	
	Notice that our Blob does not implement any of the entry interfaces that we introduces
	(Entry & BuildEntry)
*/
type BuildEntry interface {
	Entry
	GetPathname() string
	ParentDirectories() []string
}
