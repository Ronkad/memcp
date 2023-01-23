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

(print "Welcome to cpdb")

/* this can be overhooked */
(define http_handler (lambda (req res) (begin
	/* prototype req is a simple string, res is a func(string) */
	(print "new request: " req)
	((res "header") "Content-Type" "text/plain")
	((res "status") 404)
	((res "println") "404 not found")
)))

/* read  http_handler fresh from the environment */
(serve 4321 (lambda (req res) (http_handler req res)))
(print "listening on http://localhost:4321")
