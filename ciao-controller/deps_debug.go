//
// Copyright (c) 2016 Intel Corporation
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
//
// +build debug

package main

import "github.com/ciao-project/ciao/osprepare"

var controllerDeps = osprepare.PackageRequirements{
	/* sqlite3 is not strictly required, but useful for debug */
	"clearlinux": {
		{BinaryName: "/usr/bin/sqlite3", PackageName: "cloud-control"},
	},
	"fedora": {
		{BinaryName: "/usr/bin/sqlite3", PackageName: "sqlite"},
	},
	"ubuntu": {
		{BinaryName: "/usr/bin/sqlite3", PackageName: "sqlite3"},
	},
}
