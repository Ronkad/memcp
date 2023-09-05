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
package scm

import "io"
import "fmt"
import "time"
import "sync"
import "strconv"
import "net/http"
import "runtime/debug"
import "encoding/json"

// build this function into your SCM environment to offer http server capabilities
func HTTPServe(a ...Scmer) Scmer {
	// HTTP endpoint; params: (port, handler)
	port := String(a[0])
	handler := new(HttpServer)
	handler.callback = a[1] // lambda(req, res)
	server := &http.Server {
		Addr: fmt.Sprintf(":%v", port),
		Handler: handler,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	go server.ListenAndServe()
	// TODO: ListenAndServeTLS
	return "ok"
}

// HTTP handler with a scheme script underneath
type HttpServer struct {
	callback Scmer
}

func (s *HttpServer) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/plain")
	query_scm := make([]Scmer, 0)
	for k, v := range req.URL.Query() {
		for _, v2 := range v {
			query_scm = append(query_scm, k, v2)
		}
	}
	header_scm := make([]Scmer, 0)
	for k, v := range req.Header {
		for _, v2 := range v {
			header_scm = append(header_scm, k, v2)
		}
	}
	// helper
	pwtostring := func (s string, isset bool) Scmer {
		if isset {
			return s
		} else {
			return nil
		}
	}
	req_scm := []Scmer {
		"method", req.Method,
		"host", req.Host,
		"path", req.URL.Path,
		"query", query_scm,
		"header", header_scm,
		"username", req.URL.User.Username(),
		"password", pwtostring(req.URL.User.Password()),
		"ip", req.RemoteAddr,
		// TODO: req.Body io.ReadCloser
	}
	var res_lock sync.Mutex
	res_scm := []Scmer {
		"header", func (a ...Scmer) Scmer {
			res_lock.Lock()
			res.Header().Set(String(a[0]), String(a[1]))
			res_lock.Unlock();
			return "ok"
		},
		"status", func (a ...Scmer) Scmer {
			// status after header!
			res_lock.Lock()
			status, _ := strconv.Atoi(String(a[0]))
			res.WriteHeader(status)
			res_lock.Unlock();
			return "ok"
		},
		"print", func (a ...Scmer) Scmer {
			// naive output
			res_lock.Lock()
			io.WriteString(res, String(a[0]))
			res_lock.Unlock();
			return "ok"
		},
		"println", func (a ...Scmer) Scmer {
			// naive output
			res_lock.Lock()
			io.WriteString(res, String(a[0]) + "\n")
			res_lock.Unlock();
			return "ok"
		},
		"jsonl", func (a ...Scmer) Scmer {
			// print json line (only assoc)
			res_lock.Lock()
			io.WriteString(res, "{")
			dict := a[0].([]Scmer)
			for i, v := range dict {
				if i % 2 == 0 {
					// key
					//io.WriteString(res, String(a[0]) + "\n")
					bytes, _ := json.Marshal(String(v))
					res.Write(bytes)
					io.WriteString(res, ": ")
				} else {
					bytes, err := json.Marshal(v)
					if err != nil {
						panic(err)
					}
					res.Write(bytes)
					if i < len(dict)-1 {
						io.WriteString(res, ", ")
					}
				}
			}
			io.WriteString(res, "}\n")
			res_lock.Unlock();
			return "ok"
		},
	}
	// catch panics and print out 500 Internal Server Error
	defer func () {
		if r := recover(); r != nil {
			fmt.Println("request failed:", req_scm, r, string(debug.Stack()))
			res.Header().Set("Content-Type", "text/plain")
			res.WriteHeader(500)
			io.WriteString(res, "500 Internal Server Error.")
		}
	}()
	Apply(s.callback, []Scmer{req_scm, res_scm})
	// TODO: req.Body io.ReadCloser
}
