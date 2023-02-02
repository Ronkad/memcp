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
/*
	memcp smart clusterable distributed database working best on nvme

	https://pkelchte.wordpress.com/2013/12/31/scm-go/

*/
package main

import "fmt"
import "os"
import "io"
import "bufio"
import "io/ioutil"
import "path/filepath"
import "github.com/launix-de/memcp/scm"
import "github.com/launix-de/memcp/storage"

var IOEnv scm.Env

func getImport(path string) func (a ...scm.Scmer) scm.Scmer {
	return func (a ...scm.Scmer) scm.Scmer {
			filename := path + "/" + scm.String(a[0])
			// TODO: filepath.Walk for wildcards
			wd := filepath.Dir(filename)
			otherPath := scm.Env {
				scm.Vars {
					"__DIR__": path,
					"__FILE__": filename,
					"import": getImport(wd),
					"load": getLoad(wd),
				},
				&IOEnv,
				true,
			}
			bytes, err := ioutil.ReadFile(filename)
			if err != nil {
				panic(err)
			}
			return scm.EvalAll(string(bytes), &otherPath)
		}
}

func getLoad(path string) func (a ...scm.Scmer) scm.Scmer {
	return func (a ...scm.Scmer) scm.Scmer {
			filename := path + "/" + scm.String(a[0])
			if len(a) > 2 {
				file, err := os.Open(filename)
				if err != nil {
					panic(err)
				}
				splitter := bufio.NewReader(file)
				delimiter := scm.String(a[2])
				if len(delimiter) != 1 {
					panic("load delimiter must be 1 byte long")
				}
				for {
					str, err := splitter.ReadString(delimiter[0])
					if err == io.EOF {
						break // file is finished
					}
					if err != nil {
						panic(err)
					}
					// go??
					scm.Apply(a[1], []scm.Scmer{str});
				}
				//split := scm.String(a[2])
			} else {
				// read in whole
				bytes, err := ioutil.ReadFile(filename)
				if err != nil {
					panic(err)
				}
				if len(a) > 1 {
					go scm.Apply(a[1], []scm.Scmer{string(bytes)});
				} else {
					return string(bytes)
				}
			}
			return "ok"
		}
}

func main() {
	fmt.Print(`memcp Copyright (C) 2023   Carl-Philip Hänsch
    This program comes with ABSOLUTELY NO WARRANTY;
    This is free software, and you are welcome to redistribute it
    under certain conditions;
`)

	// define some IO functions (scm will not provide them since it is sandboxable)
	wd, _ := os.Getwd() // libraries are relative to working directory... is that right?
	IOEnv = scm.Env {
		scm.Vars {
			"print": func (a ...scm.Scmer) scm.Scmer {
					for _, s := range a {
						fmt.Print(scm.String(s))
					}
					fmt.Println()
					return "ok"
				},
			"import": getImport(wd),
			"load": getLoad(wd),
			"serve": scm.HTTPServe,
			"mysql": scm.MySQLServe,
		},
		&scm.Globalenv,
		true, // other defines go into Globalenv
	}
	// storage initialization
	storage.Init(scm.Globalenv)
	// scripts initialization
	scm.Eval(scm.Read("(import \"lib/main.scm\")"), &IOEnv)

	// REPL shell
	scm.Repl(&IOEnv)
}
