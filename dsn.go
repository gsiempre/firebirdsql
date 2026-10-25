/*******************************************************************************
The MIT License (MIT)

Copyright (c) 2013-2020 Hajime Nakagami

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*******************************************************************************/

package firebirdsql

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type firebirdDsn struct {
	addr    string
	dbName  string
	user    string
	passwd  string
	options map[string]string
}

var ErrDsnUserUnknown = errors.New("User unknown")

func newFirebirdDsn() *firebirdDsn {
	return &firebirdDsn{options: make(map[string]string)}
}

func parseDSN(dsns string) (*firebirdDsn, error) {

	dsn := newFirebirdDsn()

	if !strings.HasPrefix(dsns, "firebird://") {
		dsns = "firebird://" + dsns
	}
	u, err := url.Parse(dsns)
	if err != nil {
		return nil, err
	}
	if u.User == nil {
		return nil, ErrDsnUserUnknown
	}
	dsn.user = u.User.Username()
	dsn.passwd, _ = u.User.Password()
	dsn.addr = u.Host
	if !strings.ContainsRune(dsn.addr, ':') {
		dsn.addr += ":3050"
	}
	dsn.dbName = u.Path
	if !strings.ContainsRune(dsn.dbName[1:], '/') {
		dsn.dbName = dsn.dbName[1:]
	}

	//Windows Path
	if strings.ContainsRune(dsn.dbName[2:], ':') {
		dsn.dbName = dsn.dbName[1:]
	}

	m, _ := url.ParseQuery(u.RawQuery)

	var default_options = map[string]string{
		"auth_plugin_name":     "Srp256",
		"charset":              "UTF8",
		"column_name_to_lower": "false",
		"page_size":            "4096",
		"role":                 "",
		"timezone":             "",
		"wire_crypt":           "true",
	}

	for k, v := range default_options {
		values, ok := m[k]
		if ok {
			dsn.options[k] = values[0]
		} else {
			dsn.options[k] = v
		}
	}

	// Fail fast on an invalid page_size before dialing.
	if _, err := parsePageSize(dsn.options["page_size"]); err != nil {
		return nil, err
	}

	return dsn, nil
}

// validPageSizes are the page sizes some supported Firebird version accepts:
// 4096-16384 on 2.5/3.0, 4096-32768 on 4.0/5.0 and 8192-32768 on 6.0. A size
// outside the server's own range is adjusted by the server (e.g. 4096 becomes
// 8192 on 6.0).
var validPageSizes = []int{4096, 8192, 16384, 32768}

// parsePageSize parses the page_size option used by firebirdsql_createdb.
func parsePageSize(s string) (int32, error) {
	if n, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
		for _, v := range validPageSizes {
			if n == v {
				return int32(n), nil
			}
		}
	}
	return 4096, fmt.Errorf("firebirdsql: invalid page_size value %q (want one of %v)", s, validPageSizes)
}
