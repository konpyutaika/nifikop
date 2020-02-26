// Copyright 2019 Orange
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// 	You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// 	See the License for the specific language governing permissions and
// limitations under the License.

//In scratch docker container, there is no binary and we cannot stat the file
// this binary goal is to be used by the k8s healthcheck for CassKop

package main

import (
	"os"

	"github.com/sirupsen/logrus"
)

func main() {
	if _, err := os.Stat("/tmp/operator-sdk-ready"); err == nil {
		os.Exit(0)

	} else if os.IsNotExist(err) {
		logrus.Infof("error file don't exists : %v", err)
		os.Exit(1)
	} else {
		logrus.Infof("error %v", err)
		os.Exit(1)
	}
}
