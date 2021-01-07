// Copyright 2020 Celo Org
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package services

const (
	// TODO revisit: manage dependency upon RosettaCore server running alongside this
	RosettaCoreVersion = "master@c749ba869cc8fd70f4719fa726e0efead033c60b"
)

var (
	// MiddlewareVersion is the version of this package.
	// We set this as a variable instead of a constant because
	// we typically need the pointer of this // value.
	MiddlewareVersion = "0.0.1"
)
