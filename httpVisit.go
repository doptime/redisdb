package redisdb

import (
	"strings"

	cmap "github.com/orcaman/concurrent-map/v2"
)

// OpType 约束所有操作类型，保证类型安全
type OpType interface {
	~uint64
}

// -----------------------------------------------------------------------------
//  Hash
// -----------------------------------------------------------------------------

type HashOp uint64

const (
	HGet HashOp = 1 << iota
	HSet        // Value: 1<<1 (2)
	HDel        // Value: 1<<2 (4)
	HMGET
	HExists
	HGetAll
	HRandField
	HLen
	HKeys
	HVals
	HIncrBy
	HIncrByFloat
	HSetNX
	HScan

	HashRead  = HGet | HMGET | HExists | HGetAll | HRandField | HLen | HKeys | HVals | HScan
	HashWrite = HSet | HDel | HIncrBy | HIncrByFloat | HSetNX
	HashAll   = HashRead | HashWrite
)

// -----------------------------------------------------------------------------
//  List
// -----------------------------------------------------------------------------

type ListOp uint64

const (
	RPush ListOp = 1 << iota
	RPushX
	LPush
	RPop
	LPop
	LRange
	LRem
	LSet
	LIndex
	BLPop
	BRPop
	BRPopLPush
	LInsertBefore
	LInsertAfter
	Sort
	LTrim
	LLen

	ListRead  = LRange | LIndex | Sort | LLen
	ListWrite = RPush | RPushX | LPush | RPop | LPop | LRem | LSet | BLPop | BRPop | BRPopLPush | LInsertBefore | LInsertAfter | LTrim
	ListAll   = ListRead | ListWrite
)

// -----------------------------------------------------------------------------
//  Set
// -----------------------------------------------------------------------------

type SetOp uint64

const (
	SAdd SetOp = 1 << iota
	SCard
	SRem
	SIsMember
	SMembers
	SScan

	SetRead  = SCard | SIsMember | SMembers | SScan
	SetWrite = SAdd | SRem
	SetAll   = SetRead | SetWrite
)

// -----------------------------------------------------------------------------
//  ZSet
// -----------------------------------------------------------------------------

type ZSetOp uint64

const (
	ZAdd ZSetOp = 1 << iota
	ZRem
	ZRange
	ZRangeWithScores
	ZRevRangeWithScores
	ZRank
	ZRevRank
	ZScore
	ZCard
	ZCount
	ZRangeByScore
	ZRangeByScoreWithScores
	ZRevRangeByScore
	ZRevRange
	ZRevRangeByScoreWithScores
	ZRemRangeByRank
	ZRemRangeByScore
	ZIncrBy
	ZPopMax
	ZPopMin
	ZLexCount
	ZScan

	ZSetRead  = ZRange | ZRangeWithScores | ZRevRangeWithScores | ZRank | ZRevRank | ZScore | ZCard | ZCount | ZRangeByScore | ZRangeByScoreWithScores | ZRevRangeByScore | ZRevRange | ZRevRangeByScoreWithScores | ZLexCount | ZScan
	ZSetWrite = ZAdd | ZRem | ZRemRangeByRank | ZRemRangeByScore | ZIncrBy | ZPopMax | ZPopMin
	ZSetAll   = ZSetRead | ZSetWrite
)

// -----------------------------------------------------------------------------
//  String
// -----------------------------------------------------------------------------

type StringOp uint64

const (
	Get StringOp = 1 << iota
	Set
	Del
	StringGetAll
	StringSetAll

	StringRead  = Get | StringGetAll
	StringWrite = Set | Del | StringSetAll
	StringAll   = StringRead | StringWrite
)

// -----------------------------------------------------------------------------
//  Stream
// -----------------------------------------------------------------------------

type StreamOp uint64

const (
	XAdd StreamOp = 1 << iota
	XDel
	XRange
	XLen
	XRead
	XTrim
	XInfo

	StreamRead  = XRange | XLen | XRead | XInfo
	StreamWrite = XAdd | XDel | XTrim
	StreamAll   = StreamRead | StreamWrite
)

// -----------------------------------------------------------------------------
//  VectorSet Operations
// -----------------------------------------------------------------------------

type VectorSetOp uint64

const (
	FtCreate VectorSetOp = 1 << iota
	FtSearch             // Covers FT.SEARCH
	FtAggregate
	FtDropIndex
	FtAliasAdd
	FtAliasUpdate
	FtAliasDel
	FtTagVals
	FtSugAdd
	FtSugGet
	FtSugDel
	FtSugLen
	FtInfo

	// Group Masks
	VectorSetRead  = FtSearch | FtAggregate | FtTagVals | FtSugGet | FtSugLen | FtInfo
	VectorSetWrite = FtCreate | FtDropIndex | FtAliasAdd | FtAliasUpdate | FtAliasDel | FtSugAdd | FtSugDel
	VectorSetAll   = VectorSetRead | VectorSetWrite
)

type KeyOp uint64

const (
	DelKey     KeyOp = 1 << iota // 对应 DEL
	ExistsKey                    // 对应 EXISTS
	ExpireKey                    // 对应 EXPIRE, EXPIREAT
	PersistKey                   // 对应 PERSIST
	TTLKey                       // 对应 TTL, PTTL
	RenameKey                    // 对应 RENAME, RENAMEX
	TypeKey                      // 对应 TYPE
	KeysKey                      // 对应 KEYS (全局搜索，通常需要严格控制)
	TimeKey                      // 对应 TIME

	KeyRead  = ExistsKey | TTLKey | TypeKey | TimeKey
	KeyWrite = DelKey | ExpireKey | PersistKey | RenameKey
	KeyAll   = KeyRead | KeyWrite | KeysKey
)

// -----------------------------------------------------------------------------
//  Permissions Logic
// -----------------------------------------------------------------------------

var HttpPermissions = cmap.New[uint64]()

func IsAllowedHashOp(key string, op HashOp) bool {
	mask, ok := HttpPermissions.Get(keyScope(key))
	return ok && (mask&uint64(op)) != 0
}

func IsAllowedListOp(key string, op ListOp) bool {
	mask, ok := HttpPermissions.Get(keyScope(key))
	return ok && (mask&uint64(op)) != 0
}

func IsAllowedSetOp(key string, op SetOp) bool {
	mask, ok := HttpPermissions.Get(keyScope(key))
	return ok && (mask&uint64(op)) != 0
}

func IsAllowedZSetOp(key string, op ZSetOp) bool {
	mask, ok := HttpPermissions.Get(keyScope(key))
	return ok && (mask&uint64(op)) != 0
}

func IsAllowedStringOp(key string, op StringOp) bool {
	mask, ok := HttpPermissions.Get(keyScope(key))
	return ok && (mask&uint64(op)) != 0
}

func IsAllowedStreamOp(key string, op StreamOp) bool {
	mask, ok := HttpPermissions.Get(keyScope(key))
	return ok && (mask&uint64(op)) != 0
}
func IsAllowedVectorSetOp(key string, op VectorSetOp) bool {
	mask, ok := HttpPermissions.Get(keyScope(key))
	return ok && (mask&uint64(op)) != 0
}

func IsAllowedKeyOp(key string, op KeyOp) bool {
	mask, ok := HttpPermissions.Get(keyScope(key))
	return ok && (mask&uint64(op)) != 0
}
func AllowKeyOp(key string, op KeyOp) {
	mask, _ := HttpPermissions.Get(keyScope(key))
	HttpPermissions.Set(keyScope(key), mask|uint64(op))
}

func keyScope(key string) string {
	if before, _, found := strings.Cut(key, ":"); found {
		return before
	}
	return key
}
