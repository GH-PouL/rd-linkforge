package link

// Cache is the lookup cache Resolve consults before the store.
type Cache interface {
	Get(code string) (Link, bool)
	Put(code string, l Link)
	Delete(code string)
}
