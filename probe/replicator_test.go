//  Copyright (c) 2015 Rackspace
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
//  implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package probe

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	hummingbird "hummingbird/common"
)

func TestReplicationHandoff(t *testing.T) {
	e := NewEnvironment()
	defer e.Close()

	// put a file
	timestamp := hummingbird.GetTimestamp()
	assert.True(t, e.PutObject(0, timestamp, "X"))

	// make a drive look unmounted with a handler that always 507s
	origHandler := e.servers[1].Config.Handler
	e.servers[1].Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(507)
	})

	// run a primary node's replicator
	e.replicators[0].Run()

	// so it's on the primary nodes that are up
	assert.True(t, e.ObjExists(0, timestamp))
	assert.True(t, e.ObjExists(2, timestamp))

	// and now it's on the handoff node
	assert.True(t, e.ObjExists(3, timestamp))

	// fix the "unmounted" drive
	e.servers[1].Config.Handler = origHandler

	// make sure it's not on the newly fixed node yet
	assert.False(t, e.ObjExists(1, timestamp))

	// run the handoff node's replicator
	e.replicators[3].Run()

	// it's no longer on the handoff node
	assert.False(t, e.ObjExists(3, timestamp))

	// make sure it's on all the primary nodes
	assert.True(t, e.ObjExists(0, timestamp))
	assert.True(t, e.ObjExists(1, timestamp))
	assert.True(t, e.ObjExists(2, timestamp))
}