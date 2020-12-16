// Copyright 2018-2020 Burak Sezer
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package partitions

import (
	"unsafe"

	"github.com/buraksezer/olric/hasher"
	"github.com/buraksezer/olric/internal/discovery"
)

type Kind int

func (k Kind) String() string {
	if k == PRIMARY {
		return "Primary"
	} else if k == BACKUP {
		return "Backup"
	} else {
		return "Unknown"
	}
}

const (
	PRIMARY = Kind(iota + 1)
	BACKUP
)

type Partitions struct {
	count  uint64
	kind   Kind
	hasher hasher.Hasher
	m      map[uint64]*Partition
}

func New(count uint64, kind Kind, hs hasher.Hasher) *Partitions {
	ps := &Partitions{
		kind:   kind,
		count:  count,
		hasher: hs,
		m:      make(map[uint64]*Partition),
	}
	for i := uint64(0); i < count; i++ {
		ps.m[i] = &Partition{
			Id:   i,
			Kind: kind,
		}
	}
	return ps
}

// PartitionById returns the partition for the given HKey
func (ps *Partitions) PartitionById(partID uint64) *Partition {
	return ps.m[partID]
}

// PartitionIdByHKey returns partition ID for a given HKey.
func (ps *Partitions) PartitionIdByHKey(hkey uint64) uint64 {
	return hkey % ps.count
}

// PartitionByHKey returns the partition for the given HKey
func (ps *Partitions) PartitionByHKey(hkey uint64) *Partition {
	partID := ps.PartitionIdByHKey(hkey)
	return ps.m[partID]
}

// PartitionOwnersByHKey loads the partition owners list for a given hkey.
func (ps *Partitions) PartitionOwnersByHKey(hkey uint64) []discovery.Member {
	part := ps.PartitionByHKey(hkey)
	return part.owners.Load().([]discovery.Member)
}

// PartitionOwnersByHKey loads the partition owners list for a given hkey.
func (ps *Partitions) PartitionOwnersById(partID uint64) []discovery.Member {
	part := ps.PartitionById(partID)
	return part.owners.Load().([]discovery.Member)
}

// HKey returns hash-key, a.k.a hkey, for a key on a dmap.
func (ps *Partitions) HKey(name, key string) uint64 {
	tmp := name + key
	return ps.hasher.Sum64(*(*[]byte)(unsafe.Pointer(&tmp)))
}

// findPartitionOwner finds the partition Owner for a key on a dmap.
func (ps *Partitions) PartitionOwner(name, key string) (discovery.Member, uint64) {
	hkey := ps.HKey(name, key)
	return ps.PartitionByHKey(hkey).Owner(), hkey
}