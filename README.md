# Go-trie

Package Go-trie implements the prefix tree using bit operations for indexing runes. Only latin characters and/or digits are allowed for keys in the trie.

This trie implementation is very fast and has low memory usage. Also it uses mutexes so it is thread-safe and can be used in concurrent applications. 

The trie can be useful for dictionaries, search engines etc. One can store variables of any type in the trie and search for them by partial or full key match.

## Install

go get github.com/rovud/go-trie

