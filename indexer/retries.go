package indexer

import (
	"log"
	"time"

	"github.com/blockstack/blockstack.go/blockstack"
)

// GetNamesInNamespace wraps the function by the same name from blockstack.go in a retry wrapper
func (idx *Indexer) GetNamesInNamespace(ns string, offset int, count int) (blockstack.GetNamesInNamespaceResult, error) {
	return idx.retryGetNamesInNamespace(idx.retries, idx.timeout, ns, offset, count, idx.BSK.GetNamesInNamespace)
}

func (idx *Indexer) retryGetNamesInNamespace(attempts int, sleep time.Duration, ns string, offset int, count int, fn func(string, int, int) (blockstack.GetNamesInNamespaceResult, blockstack.Error)) (blockstack.GetNamesInNamespaceResult, error) {
	names, err := fn(ns, offset, count)
	if err != nil {
		if attempts--; attempts > 0 {
			time.Sleep(sleep)
			log.Printf("[blockstack] GetNamesInNamespaces failed for %s, retrying %d times\n", ns, attempts)
			return idx.retryGetNamesInNamespace(attempts, 2*sleep, ns, offset, count, fn)
		}
		return blockstack.GetNamesInNamespaceResult{}, err
	}
	return names, nil
}

// GetAllNamespaces wraps the function by the same name from blockstack.go in a retry wrapper
func (idx *Indexer) GetAllNamespaces() (blockstack.GetAllNamespacesResult, error) {
	return idx.retryGetAllNamespaces(idx.retries, idx.timeout, idx.BSK.GetAllNamespaces)
}

func (idx *Indexer) retryGetAllNamespaces(attempts int, sleep time.Duration, fn func() (blockstack.GetAllNamespacesResult, blockstack.Error)) (blockstack.GetAllNamespacesResult, error) {
	namespaces, err := fn()
	if err != nil {
		if attempts--; attempts > 0 {
			time.Sleep(sleep)
			log.Printf("[blockstack] GetAllNamespaces failed, retrying %d times\n", attempts)
			return idx.retryGetAllNamespaces(attempts, 2*sleep, fn)
		}
		return blockstack.GetAllNamespacesResult{}, err
	}
	return namespaces, nil
}

// GetNumNamesInNamespace wraps the function by the same name from blockstack.go in a retry wrapper
func (idx *Indexer) GetNumNamesInNamespace(namespace string) (blockstack.CountResult, error) {
	return idx.retryGetNumNamesInNamespace(idx.retries, idx.timeout, namespace, idx.BSK.GetNumNamesInNamespace)
}

func (idx *Indexer) retryGetNumNamesInNamespace(attempts int, sleep time.Duration, namespace string, fn func(namespace string) (blockstack.CountResult, blockstack.Error)) (blockstack.CountResult, error) {
	namespaces, err := fn(namespace)
	if err != nil {
		if attempts--; attempts > 0 {
			time.Sleep(sleep)
			log.Printf("[blockstack] GetNumNamesInNamespace for %s failed, retrying %d times\n", namespace, attempts)
			return idx.retryGetNumNamesInNamespace(attempts, 2*sleep, namespace, fn)
		}
		return blockstack.CountResult{}, err
	}
	return namespaces, nil
}

// GetZonefiles wraps the function by the same name from blockstack.go in a retry wrapper
func (idx *Indexer) GetZonefiles(zonefiles []string) (blockstack.GetZonefilesResult, error) {
	return idx.retryGetZonefiles(idx.retries, idx.timeout, zonefiles, idx.BSK.GetZonefiles)
}

func (idx *Indexer) retryGetZonefiles(attempts int, sleep time.Duration, zonefiles []string, fn func(zonefiles []string) (blockstack.GetZonefilesResult, blockstack.Error)) (blockstack.GetZonefilesResult, error) {
	zfs, err := fn(zonefiles)
	if err != nil {
		if attempts--; attempts > 0 {
			time.Sleep(sleep)
			log.Printf("[blockstack] GetZonefiles for %d zonefiles failed, retrying %d times\n", len(zonefiles), attempts)
			return idx.retryGetZonefiles(attempts, 2*sleep, zonefiles, fn)
		}
		return blockstack.GetZonefilesResult{}, err
	}
	return zfs, nil
}

// GetNameBlockchainRecord wraps the function by the same name from blockstack.go in a retry wrapper
func (idx *Indexer) GetNameBlockchainRecord(name string) (blockstack.GetNameBlockchainRecordResult, error) {
	return idx.retryGetNameBlockchainRecord(idx.retries, idx.timeout, name, idx.BSK.GetNameBlockchainRecord)
}

func (idx *Indexer) retryGetNameBlockchainRecord(attempts int, sleep time.Duration, name string, fn func(name string) (blockstack.GetNameBlockchainRecordResult, blockstack.Error)) (blockstack.GetNameBlockchainRecordResult, error) {
	zfs, err := fn(name)
	if err != nil {
		if attempts--; attempts > 0 {
			time.Sleep(sleep)
			log.Printf("[blockstack] GetNameBlockchainRecord for %s failed, retrying %d times\n", name, attempts)
			return idx.retryGetNameBlockchainRecord(attempts, 2*sleep, name, fn)
		}
		return blockstack.GetNameBlockchainRecordResult{}, err
	}
	return zfs, nil
}
