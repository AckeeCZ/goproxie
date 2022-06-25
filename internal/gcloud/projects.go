package gcloud

import "sync"

type Project struct {
	ID     string
	Name   string
	Number string
}

type projectMetadata struct {
	cacheLock               sync.Mutex
	cache                   []*Project
	cacheWarmupProgressLock sync.Mutex
	cacheWarmupInProgress   bool
}

var ProjectMetadata projectMetadata

func (pm *projectMetadata) warmupCache() {
	defer pm.cacheLock.Unlock()
	pm.cacheLock.Lock()
	if pm.cache != nil {
		return
	}
	if pm.cacheWarmupInProgress {
		return
	}
	pm.cacheWarmupProgressLock.Lock()
	pm.cacheWarmupInProgress = true
	pm.cacheWarmupProgressLock.Unlock()
	pm.cache = make([]*Project, 0)
	list := ProjectsListAllInfo()
	pm.cache = make([]*Project, len(list))
	for i, item := range list {
		pm.cache[i] = &Project{
			ID:     item[0],
			Name:   item[1],
			Number: item[2],
		}
	}
	pm.cacheWarmupProgressLock.Lock()
	pm.cacheWarmupInProgress = false
	pm.cacheWarmupProgressLock.Unlock()
}

func (pm *projectMetadata) GetProjectById(id string) *Project {
	go pm.warmupCache()
	for _, project := range pm.cache {
		if project.ID == id {
			return project
		}
	}
	return nil
}
