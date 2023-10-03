/*
Copyright (C) 2023  Carl-Philip Hänsch

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/
package storage

import "fmt"
import "sync"
import "github.com/launix-de/memcp/scm"

type dataset []scm.Scmer
type column struct {
	Name string
	Typ string
	Typdimensions []int // type dimensions for DECIMAL(10,3) and VARCHAR(5)
	Extrainfo string // TODO: further diversify into NOT NULL, AUTOINCREMENT etc.
}
type table struct {
	schema *database
	Name string
	Columns []column
	mu sync.Mutex // schema lock

	// storage
	Shards []*storageShard
	// TODO: data structure to per-column value-based shard map
	// every shard has a min-value and a max-value
	// problem: naive approach needs len(Shards)² complexity
	// solution: two shard lists: one sorted by min-value, one sorted by max-value; also every shard remembers its max-index in the min-list and min-index in max list
	// where x > value -> peek value from shard list, iterate shardlist upwards
	// where x = value -> only peek the range
	// TODO: move rows between shards when values get too widespread; triggered on equi-joins
	// the rebalance algorithm is run on a list of shards (mostly the range of shards on a equi-join; goal is to minimize the range to 1)
	// rebalance sorts all values of a column in an index (sharnr+recordid), then moves all items that would change shardid via delete+insert command
}

const max_shardsize = 65536 // dont overload the shards to get a responsive parallel full table scan

func (t *table) ShowColumns() scm.Scmer {
	result := make([]scm.Scmer, len(t.Columns))
	for i, v := range t.Columns {
		result[i] = v.Show()
	}
	return result
}

func (c *column) Show() scm.Scmer {
	dims := make([]scm.Scmer, len(c.Typdimensions))
	for i, v := range c.Typdimensions {
		dims[i] = v
	}
	return []scm.Scmer{"name", c.Name, "type", c.Typ, "dimensions", dims}
}

func (d dataset) Get(key string) scm.Scmer {
	for i := 0; i < len(d); i += 2 {
		if d[i] == key {
			return d[i+1]
		}
	}
	return nil
}

func (t *table) CreateColumn(name string, typ string, typdimensions[] int, extrainfo string) {
	t.mu.Lock()
	t.Columns = append(t.Columns, column{name, typ, typdimensions, extrainfo})
	for i := range t.Shards {
		t.Shards[i].columns[name] = new (StorageSCMER)
	}
	t.schema.save()
	t.mu.Unlock()
}

func (t *table) Insert(d dataset) {
	// load balance: if bucket is full, create new one
	shard := t.Shards[len(t.Shards)-1]
	if shard.Count() >= max_shardsize {
		t.mu.Lock()
		// reload shard after lock to avoid race conditions
		shard = t.Shards[len(t.Shards)-1]
		if shard.Count() >= max_shardsize {
			go func(i int) {
				// rebuild full shards in background
				t.Shards[i] = t.Shards[i].rebuild()
				// write new uuids to disk
				t.schema.save()
			}(len(t.Shards)-1)
			shard = NewShard(t)
			fmt.Println("started new shard for table", t.Name)
			t.Shards = append(t.Shards, shard)
		}
		t.mu.Unlock()
	}
	// physically insert
	shard.Insert(d)
}
