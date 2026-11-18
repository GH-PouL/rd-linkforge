package trie

// Trie is a prefix tree.
// It is not safe for concurrent use.
type Trie struct {
	root *node
}

type node struct {
	children map[byte]*node
	isWord   bool
}

func New() *Trie {
	return &Trie{
		root: &node{
			children: make(map[byte]*node),
		},
	}
}

func (t *Trie) Insert(word string) bool {
	if word == "" {
		return false
	}

	current := t.root

	for i := 0; i < len(word); i++ {
		ch := word[i]

		if current.children[ch] == nil {
			current.children[ch] = &node{
				children: make(map[byte]*node),
			}
		}

		current = current.children[ch]
	}

	if current.isWord {
		return false
	}

	current.isWord = true
	return true
}

func (t *Trie) Contains(word string) bool {
	if word == "" {
		return false
	}

	current := t.root

	for i := 0; i < len(word); i++ {
		ch := word[i]

		next, ok := current.children[ch]
		if !ok {
			return false
		}

		current = next
	}

	return current.isWord
}

func (t *Trie) Delete(word string) bool {
	if !t.Contains(word) {
		return false
	}

	current := t.root
	path := make([]*node, 0, len(word)+1)
	path = append(path, current)

	for i := 0; i < len(word); i++ {
		current = current.children[word[i]]
		path = append(path, current)
	}

	current.isWord = false

	for i := len(word) - 1; i >= 0; i-- {
		child := path[i+1]

		if child.isWord || len(child.children) > 0 {
			break
		}

		delete(path[i].children, word[i])
	}

	return true
}

func (t *Trie) WithPrefix(prefix string, limit int) []string {
	current := t.root

	for i := 0; i < len(prefix); i++ {
		next, ok := current.children[prefix[i]]
		if !ok {
			return nil
		}
		current = next
	}

	var result []string

	var walk func(*node, string)
	walk = func(n *node, word string) {
		if limit > 0 && len(result) >= limit {
			return
		}

		if n.isWord {
			result = append(result, word)
		}

		for ch := 0; ch < 256; ch++ {
			child, ok := n.children[byte(ch)]
			if !ok {
				continue
			}

			if limit > 0 && len(result) >= limit {
				return
			}

			walk(child, word+string(byte(ch)))
		}
	}

	walk(current, prefix)
	return result
}
