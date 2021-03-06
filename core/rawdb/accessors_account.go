// Copyright 2018 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package rawdb

import (
	"github.com/ledgerwatch/turbo-geth/common"
	"github.com/ledgerwatch/turbo-geth/common/dbutils"
	"github.com/ledgerwatch/turbo-geth/core/types/accounts"
)

// ReadAccount reading account object from multiple buckets of db
func ReadAccount(db DatabaseReader, addrHash common.Hash, acc *accounts.Account) (bool, error) {
	enc, err := db.Get(dbutils.CurrentStateBucket, addrHash[:])
	if err != nil {
		return false, err
	}
	if err = acc.DecodeForStorage(enc); err != nil {
		return false, err
	}
	root, err := db.Get(dbutils.IntermediateTrieHashBucket, dbutils.GenerateStoragePrefix(addrHash[:], acc.Incarnation))
	if err != nil {
		return false, err
	}
	if enc == nil || root == nil {
		return false, nil
	}
	acc.Root = common.BytesToHash(root)

	return true, nil
}

func WriteAccount(db DatabaseWriter, addrHash common.Hash, acc accounts.Account) error {
	value := make([]byte, acc.EncodingLengthForStorage())
	acc.EncodeForStorage(value)
	if err := db.Put(dbutils.CurrentStateBucket, addrHash[:], value); err != nil {
		return err
	}
	if err := db.Put(dbutils.IntermediateTrieHashBucket, dbutils.GenerateStoragePrefix(addrHash[:], acc.Incarnation), acc.Root.Bytes()); err != nil {
		return err
	}
	return nil
}

type DatabaseReaderDeleter interface {
	DatabaseReader
	DatabaseDeleter
}

func DeleteAccount(db DatabaseReaderDeleter, addrHash common.Hash) error {
	enc, err := db.Get(dbutils.CurrentStateBucket, addrHash[:])
	if err != nil && err.Error() != "db: key not found" {
		return err
	}
	acc := accounts.NewAccount()
	if err = acc.DecodeForStorage(enc); err != nil {
		return err
	}

	if err := db.Delete(dbutils.CurrentStateBucket, addrHash[:]); err != nil {
		return err
	}

	if err := db.Delete(dbutils.IntermediateTrieHashBucket, dbutils.GenerateStoragePrefix(addrHash[:], acc.Incarnation)); err != nil {
		return err
	}
	return nil
}
