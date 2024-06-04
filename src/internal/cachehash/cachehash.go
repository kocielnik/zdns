/*
 * ZDNS Copyright 2022 Regents of the University of Michigan
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not
 * use this file except in compliance with the License. You may obtain a copy
 * of the License at http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
 * implied. See the License for the specific language governing
 * permissions and limitations under the License.
 */

package cachehash

import (
	"container/list"
	"sync"

	log "github.com/sirupsen/logrus"
)

type CacheHash struct {
	sync.Mutex
	h       map[interface{}]*list.Element
	l       *list.List
	len     int
	maxLen  int
	ejectCB func(interface{}, interface{})
}

type keyValue struct {
	Key   interface{}
	Value interface{}
}

func (c *CacheHash) Init(maxLen int) {
	c.l = list.New()
	c.l = c.l.Init()
	c.h = make(map[interface{}]*list.Element)
	c.len = 0
	c.maxLen = maxLen
}

func (c *CacheHash) Eject() {
	if c.len == 0 {
		// nothing to eject
		return
	}
	e := c.l.Back()
	kv, ok := e.Value.(keyValue)
	if !ok {
		log.Panic("CacheHash: Eject: invalid list element value type")
	}
	if c.ejectCB != nil {
		c.ejectCB(kv.Key, kv.Value)
	}
	delete(c.h, kv.Key)
	c.l.Remove(e)
	c.len--
}

// Add adds a key-value pair to the cache.
// If the key already exists, the value is updated and the element is moved to the front of the list.
// Returns true if the key already existed in the cache, false otherwise.
func (c *CacheHash) Add(k interface{}, v interface{}) bool {
	e, elemFound := c.h[k]
	if elemFound {
		kv, ok := e.Value.(keyValue)
		if !ok {
			log.Panic("CacheHash: Add: invalid list element value type")
		}
		kv.Key = k
		kv.Value = v
		c.l.MoveToFront(e)
		return true
	}
	if c.len >= c.maxLen {
		// cache is full, eject the least-used element
		c.Eject()
	}
	var kv keyValue
	kv.Key = k
	kv.Value = v
	e = c.l.PushFront(kv)
	c.len++
	c.h[k] = e
	return false
}

func (c *CacheHash) First() (interface{}, interface{}) {
	if c.len == 0 {
		return nil, nil
	}
	e := c.l.Front()
	kv, ok := e.Value.(keyValue)
	if !ok {
		log.Panic("CacheHash: First: invalid list element value type")
	}
	return kv.Key, kv.Value
}

func (c *CacheHash) Last() (interface{}, interface{}) {
	if c.len == 0 {
		return nil, nil
	}
	e := c.l.Back()
	kv, ok := e.Value.(keyValue)
	if !ok {
		log.Panic("CacheHash: Last: invalid list element value type")
	}
	return kv.Key, kv.Value
}

func (c *CacheHash) Get(k interface{}) (interface{}, bool) {
	e, ok := c.h[k]
	if !ok {
		return nil, false
	}
	c.l.MoveToFront(e)
	kv, ok := e.Value.(keyValue)
	if !ok {
		log.Panic("CacheHash: Get: invalid list element value type")
	}
	return kv.Value, true
}

func (c *CacheHash) GetNoMove(k interface{}) (interface{}, bool) {
	e, ok := c.h[k]
	if !ok {
		return nil, false
	}
	kv, ok := e.Value.(keyValue)
	if !ok {
		log.Panic("CacheHash: GetNoMove: invalid list element value type")
	}
	return kv.Value, true
}

func (c *CacheHash) Has(k interface{}) bool {
	_, ok := c.h[k]
	return ok
}

func (c *CacheHash) Delete(k interface{}) (interface{}, bool) {
	e, ok := c.h[k]
	if !ok {
		return nil, false
	}
	kv, ok := e.Value.(keyValue)
	if !ok {
		log.Panic("CacheHash: Delete: invalid list element value type")
	}
	delete(c.h, k)
	c.l.Remove(e)
	c.len--
	return kv.Value, true
}

func (c *CacheHash) Len() int {
	return c.len
}

func (c *CacheHash) RegisterCB(newCB func(interface{}, interface{})) {
	c.ejectCB = newCB
}
