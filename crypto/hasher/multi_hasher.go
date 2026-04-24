// Copyright (C) 2025 T-Force I/O
// This file is part of TFunifiler
//
// TFunifiler is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// TFunifiler is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with TFunifiler. If not, see <https://www.gnu.org/licenses/>.

package hasher

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"fmt"
	"hash"
	"hash/crc32"
	"io"
	"os"

	"github.com/tforceaio/tf-unifiler/diag"
	"golang.org/x/crypto/md4"
	"golang.org/x/crypto/ripemd160"
)

// HashResult stores the checksum of a file for specific algorithm.
type HashResult struct {
	Path      string
	Size      int
	Algorithm string
	Hash      []byte
}

// Compute multiple checksums of a file.
func Hash(fPath string, algorithms []string) ([]*HashResult, error) {
	fHandle, err := os.Open(fPath)
	if err != nil {
		return []*HashResult{}, err
	}
	stat, err := fHandle.Stat()
	if err != nil {
		fHandle.Close()
		return []*HashResult{}, err
	}
	defer fHandle.Close()

	results := make([]*HashResult, len(algorithms))
	hashers := make([]hash.Hash, len(algorithms))
	for i, a := range algorithms {
		results[i] = &HashResult{
			Path:      fPath,
			Algorithm: a,
		}
		switch a {
		case "crc32":
			hashers[i] = crc32.NewIEEE()
		case "crc32c":
			hashers[i] = crc32.New(crc32.MakeTable(crc32.Castagnoli))
		case "md4":
			hashers[i] = md4.New()
		case "md5":
			hashers[i] = md5.New()
		case "ripemd160":
			hashers[i] = ripemd160.New()
		case "sha1":
			hashers[i] = sha1.New()
		case "sha224":
			hashers[i] = sha256.New224()
		case "sha256":
			hashers[i] = sha256.New()
		case "sha384":
			hashers[i] = sha512.New384()
		case "sha512":
			hashers[i] = sha512.New()
		default:
			return []*HashResult{}, fmt.Errorf("unsupported hash algorithm: '%s'", a)
		}
	}

	p := diag.NewProgressTracker("Hash File", notifier)
	defer p.Done()
	p.Total(stat.Size())
	p.Status(fPath)

	bufSize := getBufferSize(fHandle)
	buf := make([]byte, bufSize)
	written := int64(0)
	for {
		nread, eread := fHandle.Read(buf)
		if nread > 0 {
			nwrite := 0
			var ewrite error = nil
			for _, hasher := range hashers {
				nwrite, ewrite = hasher.Write(buf[0:nread])
			}
			if nwrite < 0 || nread < nwrite {
				nwrite = 0
				if ewrite == nil {
					ewrite = errors.New("cannot write to hasher")
				}
			}
			written += int64(nwrite)
			p.Progress(written)
			if ewrite != nil {
				err = ewrite
				break
			}
			if nread != nwrite {
				err = fmt.Errorf("read and write data mismatch %d %d", nread, nwrite)
				break
			}
		}
		if eread != nil {
			if eread != io.EOF {
				err = eread
			}
			break
		}
	}
	if err != nil {
		return []*HashResult{}, err
	}

	for i, h := range hashers {
		results[i].Size = int(written)
		results[i].Hash = h.Sum(nil)
	}
	return results, nil
}

func getBufferSize(src io.Reader) int {
	size := 32 * 1024
	if l, ok := src.(*io.LimitedReader); ok && int64(size) > l.N {
		if l.N < 1 {
			size = 1
		} else {
			size = int(l.N)
		}
	}
	return size
}

// Compute checksum of a file.
func hashFile(fPath string, hasher hash.Hash, algo string) (*HashResult, error) {
	fHandle, err := os.Open(fPath)
	if err != nil {
		return nil, err
	}
	stat, err := fHandle.Stat()
	if err != nil {
		return nil, err
	}
	defer fHandle.Close()

	p := diag.NewProgressTracker("Hash File", notifier)
	defer p.Done()
	p.Total(stat.Size())
	p.Status(fPath)

	bufSize := getBufferSize(fHandle)
	buf := make([]byte, bufSize)
	written := int64(0)
	for {
		nread, eread := fHandle.Read(buf)
		if nread > 0 {
			nwrite, ewrite := hasher.Write(buf[0:nread])
			if nwrite < 0 || nread < nwrite {
				nwrite = 0
				if ewrite == nil {
					ewrite = errors.New("cannot write to hasher")
				}
			}
			written += int64(nwrite)
			p.Progress(written)
			if ewrite != nil {
				err = ewrite
				break
			}
			if nread != nwrite {
				err = fmt.Errorf("read and write data mismatch %d %d", nread, nwrite)
				break
			}
		}
		if eread != nil {
			if eread != io.EOF {
				err = eread
			}
			break
		}
	}
	if err != nil {
		return nil, err
	}

	result := &HashResult{
		Path:      fPath,
		Size:      int(written),
		Algorithm: algo,
		Hash:      hasher.Sum(nil),
	}
	return result, nil
}
