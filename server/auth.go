// Copyright (C) 2025  T-Force I/O
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package server

import (
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// ErrUnauthorized reports a request without a valid API key.
var ErrUnauthorized = errors.New("missing or invalid API key")

// Return a middleware that rejects every request not carrying "Authorization: Bearer <apiKey>".
func requireAPIKey(apiKey string) gin.HandlerFunc {
	// Compare digests so the comparison takes the same time whatever the length of the presented key.
	want := sha256.Sum256([]byte(apiKey))
	return func(c *gin.Context) {
		scheme, token, _ := strings.Cut(c.GetHeader("Authorization"), " ")
		got := sha256.Sum256([]byte(token))
		if !strings.EqualFold(scheme, "Bearer") || subtle.ConstantTimeCompare(got[:], want[:]) != 1 {
			c.Header("WWW-Authenticate", "Bearer")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": ErrUnauthorized.Error()})
			return
		}
		c.Next()
	}
}
