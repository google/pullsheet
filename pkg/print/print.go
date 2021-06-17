// Copyright 2021 Google LLC
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

package print

import (
	"encoding/json"
	"fmt"

	"github.com/gocarina/gocsv"
	"k8s.io/klog"
)

//This method prints the values in "data" interface to standatrd output in the format specified by "out_type", either JSON/CSV
func Print(data interface{}, out_type string) error {
	var (
		err error
		out string
	)

	if out_type == "JSON" {
		var jsonvar []byte
		jsonvar, err = json.Marshal(data)
		out = string(jsonvar)
	} else if out_type == "CSV" {
		out, err = gocsv.MarshalString(data)
	}

	if err != nil {
		return err
	}

	klog.Infof("%d bytes of reviews output", len(out))
	fmt.Print(out)

	return nil
}
