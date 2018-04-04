package cmd

import "fmt"

// GetNSInfo returns information about all the namespaces
func (idx *Indexer) GetNSInfo() (NSInfo, error) {
	out := make(map[string]int, 0)

	// Fetch the list of namespaces
	ns, err := idx.BSK.GetAllNamespaces()
	if err != nil {
		return nil, err
	}

	// Record this number in stats map
	idx.Record("namespaces", len(ns.Namespaces))

	// Fetch the number of names in each namespace
	for _, namespace := range ns.Namespaces {
		num, err := idx.BSK.GetNumNamesInNamespace(namespace)
		if err != nil {
			return nil, err
		}
		idx.Record(fmt.Sprintf("namespace/%s", namespace), num.Count)
		out[namespace] = num.Count
	}
	return NSInfo(out), nil
}

// NSInfo contains information about all the namespace in Blockstack
type NSInfo map[string]int

// Namespaces returns the list of namespaces
func (ns NSInfo) Namespaces() []string {
	out := []string{}
	for k := range ns {
		out = append(out, k)
	}
	return out
}

// Count returns the total count of all names in all namespaces
func (ns NSInfo) Count() int {
	out := 0
	for _, v := range ns {
		out += v
	}
	return out
}

// Pages returns the number of pages of names in a namespace
func (ns NSInfo) Pages(namespace string) int {
	return ns[namespace]/namePageSize + 2
}
