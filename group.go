// Copyright 2017 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package fn

// Group represents a handler group that contains same hooks
type Group struct {
	plugins []PluginFunc
}

func NewGroup() *Group {
	return &Group{}
}

func (g *Group) Plugin(plugins ...PluginFunc) *Group {
	for _, b := range plugins {
		if b != nil {
			g.plugins = append(g.plugins, b)
		}
	}
	return g
}

func (g *Group) Wrap(f interface{}) *fn {
	n := Wrap(f)
	if length := len(g.plugins); length > 0 {
		n.plugins = make([]PluginFunc, length)
		copy(n.plugins, g.plugins)
	}
	return n
}
