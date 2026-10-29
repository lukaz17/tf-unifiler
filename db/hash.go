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

package db

import (
	"encoding/hex"

	"github.com/google/uuid"
	"github.com/tforceaio/tf-unifiler/core"
)

// Hash represents a set of hashes for a particular file along with basic metadata.
type Hash struct {
	ID     Bytes32 `gorm:"column:id;primaryKey"`
	Md5    string  `gorm:"column:md5"`
	Sha1   string  `gorm:"column:sha1"`
	Sha256 string  `gorm:"column:sha256;uniqueIndex"`
	Sha512 string  `gorm:"column:sha512"`
	Crc32  string  `gorm:"column:crc32"`

	Size        int64  `gorm:"column:size"`
	Description string `gorm:"column:description"`
	IsIgnored   bool   `gorm:"column:is_ignored"`

	SessionID uuid.UUID `gorm:"session_id"`
}

// Return new hash.
func NewHash(fileHashes *core.FileMultiHash, isIgnored bool) *Hash {
	var id Bytes32
	copy(id[:], fileHashes.Sha256)
	return &Hash{
		ID:          id,
		Crc32:       fileHashes.Crc32.HexStr(),
		Md5:         fileHashes.Md5.HexStr(),
		Sha1:        fileHashes.Sha1.HexStr(),
		Sha256:      fileHashes.Sha256.HexStr(),
		Sha512:      fileHashes.Sha512.HexStr(),
		Size:        fileHashes.Size,
		Description: fileHashes.FileName,
		IsIgnored:   isIgnored,
	}
}

// Get Hash by ID.
func (c *DbContext) GetHash(id Bytes32) (*Hash, error) {
	return c.findHash(id)
}

// Get Hash by its SHA-256.
func (c *DbContext) GetHashBySha256(hash string) (*Hash, error) {
	return c.findHashBySha256(hash)
}

// Get Hashes belong to Sets by SetIDs.
func (c *DbContext) GetHashesBySetIDs(setIDs uuid.UUIDs) ([]*Hash, error) {
	return c.findHashesBySetIDs(setIDs)
}

// Get Hashes belong to Sets by Set Names and their SHA-256s.
func (c *DbContext) GetHashesInSets(sets, sha256s []string, onlyIgnored bool) ([]*Hash, error) {
	return c.findHashesInSets(sets, sha256s, onlyIgnored)
}

// Get Hashes by their SHA-256s.
func (c *DbContext) GetHashesBySha256s(hashes []string) ([]*Hash, error) {
	return c.findHashesBySha256s(hashes)
}

// Save Hash to database.
func (c *DbContext) SaveHash(hash *Hash) error {
	changedHash, err := c.findHashBySha256(hash.Sha256)
	if err != nil {
		return err
	}
	newHashes := []*Hash{}
	changedHashes := []*Hash{}
	if changedHash == nil {
		newHashes = append(newHashes, hash)
	} else {
		changedHashes = append(changedHashes, hash)
	}
	return c.writeHashes(newHashes, changedHashes)
}

// Save Hashes to database.
func (c *DbContext) SaveHashes(hashes []*Hash) error {
	sha256s := make([]string, len(hashes))
	for i, hash := range hashes {
		sha256s[i] = hash.Sha256
	}
	changedHashes, err := c.findHashesBySha256s(sha256s)
	if err != nil {
		return err
	}
	changedHashesMap := map[string]Bytes32{}
	for _, hash := range changedHashes {
		changedHashesMap[hash.Sha256] = hash.ID
	}
	newHashes := []*Hash{}
	for _, hash := range hashes {
		if _, ok := changedHashesMap[hash.Sha256]; ok {
			continue
		}
		newHashes = append(newHashes, hash)
		changedHashesMap[hash.Sha256] = hash.ID
	}
	return c.writeHashes(newHashes, []*Hash{})
}

// Return Hash that has specified id.
func (c *DbContext) findHash(id Bytes32) (*Hash, error) {
	var doc *Hash
	result := c.db.Model(&Hash{}).
		Where("id = ?", id).
		First(&doc)
	if c.isEmptyResultError(result.Error) {
		return nil, nil
	}
	return doc, result.Error
}

// Return Hash that has specified SHA-256.
func (c *DbContext) findHashBySha256(hash string) (*Hash, error) {
	var doc *Hash
	result := c.db.Model(&Hash{}).
		Where("sha256 = ?", hash).
		First(&doc)
	if c.isEmptyResultError(result.Error) {
		return nil, nil
	}
	return doc, result.Error
}

// Return Hashes that belong to Sets that have specified setIDs.
func (c *DbContext) findHashesBySetIDs(setIDs uuid.UUIDs) ([]*Hash, error) {
	var docs []*Hash
	result := c.db.Model(&Hash{}).
		InnerJoins("hashes ON hashes.id = set_hashes.hash_id AND set_hashes.set_id IN ?", setIDs).
		Find(&docs)
	return docs, result.Error
}

// Return Hashes that belong to Sets that have specified setNames and have specified SHA-256s.
func (c *DbContext) findHashesInSets(setNames, sha256s []string, onlyIgnored bool) ([]*Hash, error) {
	var docs []*Hash
	result := c.db.Model(&Hash{}).
		Joins("JOIN set_hashes ON hashes.id = set_hashes.hash_id").
		Joins("JOIN sets ON set_hashes.set_id = sets.id").
		Where("(0 = ? OR sets.name IN ?) AND (0 = ? OR hashes.sha256 IN ?) AND (? OR hashes.is_ignored)",
			len(setNames), setNames,
			len(sha256s), sha256s,
			!onlyIgnored, onlyIgnored,
		).
		Find(&docs)
	return docs, result.Error
}

// Return Hashes that have specified SHA-256s.
func (c *DbContext) findHashesBySha256s(hashes []string) ([]*Hash, error) {
	var docs []*Hash
	result := c.db.Model(&Hash{}).
		Where("sha256 IN ?", hashes).
		Find(&docs)
	return docs, result.Error
}

// Insert new Hashes and update old Hashes in one transaction.
func (c *DbContext) writeHashes(newHashes []*Hash, changedHashes []*Hash) error {
	tx := c.db.Begin()
	for _, hash := range newHashes {
		if hash.ID == (Bytes32{}) {
			b, err := hex.DecodeString(hash.Sha256)
			if err != nil {
				return err
			}
			copy(hash.ID[:], b)
		}
		result := tx.Create(hash)
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}
	}
	for _, hash := range changedHashes {
		result := tx.Model(&Hash{}).
			Where("id = ?", hash.ID).
			Updates(map[string]interface{}{
				"md5":         hash.Md5,
				"sha1":        hash.Sha1,
				"sha256":      hash.Sha256,
				"sha512":      hash.Sha512,
				"crc32":       hash.Crc32,
				"size":        hash.Size,
				"description": hash.Description,
				"is_ignored":  hash.IsIgnored,
				"session_id":  hash.SessionID,
			})
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}
	}
	tx.Commit()
	return nil
}

// Update is_ignored flag of Hashes identified by their SHA-256s.
// hashesToIgnore will have is_ignored set to true.
// hashesToApprove will have is_ignored set to false.
func (c *DbContext) setIgnoredBySha256s(hashesToIgnore, hashesToApprove []string) error {
	tx := c.db.Begin()
	if len(hashesToIgnore) > 0 {
		result := tx.Model(&Hash{}).
			Where("sha256 IN ?", hashesToIgnore).
			Updates(map[string]interface{}{
				"is_ignored": true,
			})
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}
	}
	if len(hashesToApprove) > 0 {
		result := tx.Model(&Hash{}).
			Where("sha256 IN ?", hashesToApprove).
			Updates(map[string]interface{}{
				"is_ignored": false,
			})
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}
	}
	return nil
}
