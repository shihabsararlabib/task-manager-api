//go:build ignore
// +build ignore

package middleware
package middleware














}	})		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))		next.ServeHTTP(w, r)		start := time.Now()	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {func Logging(next http.Handler) http.Handler {)	"time"	"net/http"	"log"import (