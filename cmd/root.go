// Copyright © 2018 NAME HERE geoffrey.ragot@gmail.com
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
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"errors"
	"io"
	"archive/zip"
	"path/filepath"
)

func copyFile(archive *zip.Writer, dst, src string) error {
	w, err := archive.Create(filepath.Base(dst))
	if err != nil {
		return err
	}
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(w, f)
	return err
}

func copyDir(archive *zip.Writer, dst, src string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		w, err := archive.Create(dst + path[len(src):])
		if err != nil {
			return err
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(w, f)
		return err
	})
}

var RootCmd = &cobra.Command{
	Use:   "binzip <static files...> <binary> <output>",
	Short: "Pack binary and files together",
	RunE: func(cmd *cobra.Command, args []string) error {

		if len(args) < 3 {
			return errors.New("Need at least 3 arguments")
		}

		assets, binary, output := args[:len(args)-2], args[len(args)-2], args[len(args)-1]

		bin, err := os.Open(binary)
		if err != nil {
			return err
		}
		defer bin.Close()

		out, err := os.Create(output)
		if err != nil {
			return err
		}
		defer out.Close()

		size, err := io.Copy(out, bin)
		if err != nil {
			return err
		}

		archive := zip.NewWriter(out)
		archive.SetOffset(size)
		defer archive.Close()

		for _, asset := range assets {
			asset := filepath.Clean(asset)
			asset, err = filepath.Abs(asset)
			if err != nil {
				return err
			}

			stat, err := os.Stat(asset)
			if err != nil {
				return err
			}

			if stat.IsDir() {
				copyDir(archive, filepath.Base(asset), asset)
			} else {
				copyFile(archive, filepath.Base(asset), asset)
			}
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}