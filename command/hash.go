// Copyright (c) 2020, Rod Dong <rod.dong@gmail.com>
// All rights reserved.
//
// Use of this source code is governed by The MIT License.

// Package command is to handle the command from client.
package command

import (
	"strconv"

	"github.com/rod6/rodis/resp"
)

// command
// ------------
// HDEL
// HEXISTS
// HGET
// HGETALL
// HINCRBY
// HINCRBYFLOAT
// HKEYS
// HLEN
// HMGET
// HMSET
// HSCAN
// HSET
// HSETNX
// HSTRLEN
// HVALS

// hdel -> https://redis.io/commands/hdel
func hdel(v Args, ex *Extras) error {
	if len(v) < 2 {
		return resp.NewError(ErrFmtWrongNumberArgument, "hdel").WriteTo(ex.Buffer)
	}

	ex.DB.Lock()
	defer ex.DB.Unlock()

	keyExists, tipe := ex.DB.Has(v[0])
	if !keyExists {
		return resp.ZeroInteger.WriteTo(ex.Buffer)
	}
	if keyExists && tipe != resp.Hash {
		return resp.NewError(ErrWrongType).WriteTo(ex.Buffer)
	}

	fields := [][]byte{}
	for _, field := range v[1:] {
		fields = append(fields, []byte(field))
	}
	hash := ex.DB.GetFields(v[0], fields)

	count := 0
	for _, value := range hash {
		if len(value) != 0 {
			count++
		}
	}
	ex.DB.DeleteFields(v[0], fields)
	return resp.Integer(count).WriteTo(ex.Buffer)
}

// hexists -> https://redis.io/commands/hexist
func hexists(v Args, ex *Extras) error {
	ex.DB.RLock()
	defer ex.DB.RUnlock()

	keyExists, tipe := ex.DB.Has(v[0])
	if keyExists && tipe != resp.Hash {
		return resp.NewError(ErrWrongType).WriteTo(ex.Buffer)
	}

	hash := ex.DB.GetFields(v[0], [][]byte{v[1]})
	if len(hash[string(v[1])]) == 0 {
		return resp.ZeroInteger.WriteTo(ex.Buffer)
	}
	return resp.OneInteger.WriteTo(ex.Buffer)
}

// hget -> https://redis.io/commands/hget
func hget(v Args, ex *Extras) error {
	ex.DB.RLock()
	defer ex.DB.RUnlock()

	keyExists, tipe := ex.DB.Has(v[0])
	if !keyExists {
		return resp.NilBulkString.WriteTo(ex.Buffer)
	}
	if keyExists && tipe != resp.Hash {
		return resp.NewError(ErrWrongType).WriteTo(ex.Buffer)
	}

	hash := ex.DB.GetFields(v[0], [][]byte{v[1]})
	if len(hash[string(v[1])]) == 0 {
		return resp.NilBulkString.WriteTo(ex.Buffer)
	}

	return resp.BulkString(hash[string(v[1])]).WriteTo(ex.Buffer)
}

// hgetall -> https://redis.io/commands/hgetall
func hgetall(v Args, ex *Extras) error {
	ex.DB.RLock()
	defer ex.DB.RUnlock()

	keyExists, tipe := ex.DB.Has(v[0])
	if !keyExists {
		return resp.EmptyArray.WriteTo(ex.Buffer)
	}
	if keyExists && tipe != resp.Hash {
		return resp.NewError(ErrWrongType).WriteTo(ex.Buffer)
	}

	hash := ex.DB.GetHashAsArray(v[0])
	arr := resp.Array{}

	for _, field := range hash {
		arr = append(arr, resp.BulkString(field.Key), resp.BulkString(field.Value))
	}
	return arr.WriteTo(ex.Buffer)
}

// hincrby -> https://redis.io/commands/hincrby
func hincrby(v Args, ex *Extras) error {
	by, err := strconv.ParseInt(string(v[2]), 10, 64)
	if err != nil {
		return resp.NewError(ErrNotValidInt).WriteTo(ex.Buffer)
	}

	ex.DB.Lock()
	defer ex.DB.Unlock()

	keyExists, tipe := ex.DB.Has(v[0])
	if keyExists && tipe != resp.Hash {
		return resp.NewError(ErrWrongType).WriteTo(ex.Buffer)
	}

	hash := ex.DB.GetFields(v[0], [][]byte{v[1]})

	newVal := int64(0)
	if len(hash[string(v[1])]) == 0 {
		newVal += by
	} else {
		i, err := strconv.ParseInt(string(hash[string(v[1])]), 10, 64)
		if err != nil {
			return resp.NewError(ErrNotValidInt).WriteTo(ex.Buffer)
		}
		newVal = i + by
	}
	hash[string(v[1])] = []byte(strconv.FormatInt(newVal, 10))

	ex.DB.PutHash(v[0], resp.Hash, hash)
	return resp.Integer(newVal).WriteTo(ex.Buffer)
}

// hincrbyfloat -> https://redis.io/commands/hincrbyfloat
func hincrbyfloat(v Args, ex *Extras) error {
	by, err := strconv.ParseFloat(string(v[2]), 64)
	if err != nil {
		return resp.NewError(ErrNotValidInt).WriteTo(ex.Buffer)
	}

	ex.DB.Lock()
	defer ex.DB.Unlock()

	exist, tipe := ex.DB.Has(v[0])
	if exist && tipe != resp.Hash {
		return resp.NewError(ErrWrongType).WriteTo(ex.Buffer)
	}

	hash := ex.DB.GetFields(v[0], [][]byte{v[1]})

	newVal := 0.0
	if len(hash[string(v[1])]) == 0 {
		newVal += by
	} else {
		f, err := strconv.ParseFloat(string(hash[string(v[1])]), 64)
		if err != nil {
			return resp.NewError(ErrNotValidFloat).WriteTo(ex.Buffer)
		}
		newVal = f + by
	}
	hash[string(v[1])] = []byte(strconv.FormatFloat(newVal, 'f', -1, 64))

	ex.DB.PutHash(v[0], resp.Hash, hash)
	return resp.BulkString(hash[string(v[1])]).WriteTo(ex.Buffer)
}

// hkeys -> https://redis.io/commands/hkeys
func hkeys(v Args, ex *Extras) error {
	ex.DB.RLock()
	defer ex.DB.RUnlock()

	keyExists, tipe := ex.DB.Has(v[0])
	if !keyExists {
		return resp.EmptyArray.WriteTo(ex.Buffer)
	}
	if keyExists && tipe != resp.Hash {
		return resp.NewError(ErrWrongType).WriteTo(ex.Buffer)
	}

	fields := ex.DB.GetFieldNames(v[0])
	arr := resp.Array{}

	for _, field := range fields {
		arr = append(arr, resp.BulkString(field))
	}
	return arr.WriteTo(ex.Buffer)
}

// hvals -> https://redis.io/commands/hvals
func hvals(v Args, ex *Extras) error {
	ex.DB.RLock()
	defer ex.DB.RUnlock()

	keyExists, tipe := ex.DB.Has(v[0])
	if !keyExists {
		return resp.EmptyArray.WriteTo(ex.Buffer)
	}
	if keyExists && tipe != resp.Hash {
		return resp.NewError(ErrWrongType).WriteTo(ex.Buffer)
	}

	hash := ex.DB.GetHashAsArray(v[0])
	arr := resp.Array{}

	for _, field := range hash {
		arr = append(arr, resp.BulkString(field.Value))
	}
	return arr.WriteTo(ex.Buffer)
}

// hlen -> https://redis.io/commands/hlen
func hlen(v Args, ex *Extras) error {
	ex.DB.RLock()
	defer ex.DB.RUnlock()

	keyExists, tipe := ex.DB.Has(v[0])
	if !keyExists {
		return resp.ZeroInteger.WriteTo(ex.Buffer)
	}
	if keyExists && tipe != resp.Hash {
		return resp.NewError(ErrWrongType).WriteTo(ex.Buffer)
	}

	fields := ex.DB.GetFieldNames(v[0])
	return resp.Integer(len(fields)).WriteTo(ex.Buffer)
}

// hmget -> https://redis.io/commands/hmget
func hmget(v Args, ex *Extras) error {
	if len(v) < 2 {
		return resp.NewError(ErrFmtWrongNumberArgument, "hmget").WriteTo(ex.Buffer)
	}

	ex.DB.RLock()
	defer ex.DB.RUnlock()

	keyExists, tipe := ex.DB.Has(v[0])
	if keyExists && tipe != resp.Hash {
		return resp.NewError(ErrWrongType).WriteTo(ex.Buffer)
	}

	fields := [][]byte{}
	for _, f := range v[1:] {
		fields = append(fields, f)
	}
	hash := ex.DB.GetFieldsAsArray(v[0], fields)

	arr := resp.Array{}
	for _, field := range hash {
		if len(field.Value) == 0 {
			arr = append(arr, resp.NilBulkString)
		} else {
			arr = append(arr, resp.BulkString(field.Value))
		}
	}
	return arr.WriteTo(ex.Buffer)
}

// hmset -> https://redis.io/commands/hmset
func hmset(v Args, ex *Extras) error {
	if len(v) <= 1 || len(v)%2 != 1 {
		return resp.NewError(ErrFmtWrongNumberArgument, "hmset").WriteTo(ex.Buffer)
	}

	ex.DB.Lock()
	defer ex.DB.Unlock()

	exist, tipe := ex.DB.Has(v[0])
	if exist && tipe != resp.Hash {
		return resp.NewError(ErrWrongType).WriteTo(ex.Buffer)
	}

	hash := make(map[string][]byte)
	for i := 1; i < len(v); {
		hash[string(v[i])] = v[i+1]
		i += 2
	}
	ex.DB.PutHash(v[0], resp.Hash, hash)
	return resp.OkSimpleString.WriteTo(ex.Buffer)
}

// hset -> https://redis.io/commands/hset
func hset(v Args, ex *Extras) error {
	ex.DB.Lock()
	defer ex.DB.Unlock()

	exist, tipe := ex.DB.Has(v[0])
	if exist && tipe != resp.Hash {
		return resp.NewError(ErrWrongType).WriteTo(ex.Buffer)
	}

	fieldExists := false
	hash := ex.DB.GetFields(v[0], [][]byte{v[1]})
	if len(hash[string(v[1])]) != 0 {
		fieldExists = true
	}

	hash[string(v[1])] = v[2]
	ex.DB.PutHash(v[0], resp.Hash, hash)

	if !fieldExists {
		return resp.OneInteger.WriteTo(ex.Buffer)
	}
	return resp.ZeroInteger.WriteTo(ex.Buffer)
}

// hsetnx -> https://redis.io/commands/hsetnx
func hsetnx(v Args, ex *Extras) error {
	ex.DB.Lock()
	defer ex.DB.Unlock()

	exist, tipe := ex.DB.Has(v[0])
	if exist && tipe != resp.Hash {
		return resp.NewError(ErrWrongType).WriteTo(ex.Buffer)
	}

	fieldExists := false
	hash := ex.DB.GetFields(v[0], [][]byte{v[1]})
	if len(hash[string(v[1])]) != 0 {
		fieldExists = true
	}

	if !fieldExists {
		hash[string(v[1])] = v[2]
		ex.DB.PutHash(v[0], resp.Hash, hash)
		return resp.OneInteger.WriteTo(ex.Buffer)
	}
	return resp.ZeroInteger.WriteTo(ex.Buffer)
}

// hstrlen -> https://redis.io/commands/hstrlen
func hstrlen(v Args, ex *Extras) error {
	ex.DB.RLock()
	defer ex.DB.RUnlock()

	exist, tipe := ex.DB.Has(v[0])
	if !exist {
		return resp.ZeroInteger.WriteTo(ex.Buffer)
	}
	if exist && tipe != resp.Hash {
		return resp.NewError(ErrWrongType).WriteTo(ex.Buffer)
	}

	hash := ex.DB.GetFields(v[0], [][]byte{v[1]})
	return resp.Integer(len(hash[string(v[1])])).WriteTo(ex.Buffer)
}
