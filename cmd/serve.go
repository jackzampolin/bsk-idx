// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
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

package cmd

import (
	"bytes"
	"encoding/gob"

	"github.com/jackzampolin/bsk-idx/indexer"
	"github.com/spf13/cobra"
)

func encodeStringArray(strs []string) []byte {
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	enc.Encode(strs)
	return buf.Bytes()
}

func decodeStringArray(byt []byte) []string {
	return []string{}
}

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		idx := indexer.NewIndexer(cfg, []string{})
		idx.Index()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
