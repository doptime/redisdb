package redisdb

import (
	"strings"

	"github.com/doptime/logger"
	cmap "github.com/orcaman/concurrent-map/v2"
)

// SystemDbKey 是全局/系统级操作（如 TIME, KEYS）的专用鉴权键
const SystemDbKey = "_systemdb"

// -----------------------------------------------------------------------------
//  1. 基础通用权限 (无类型常量)
//  位区间：0-9。全局通用，任何类型自动适配。
// -----------------------------------------------------------------------------

const (
	Del     = 1 << 0 // 对应 DEL
	Exists  = 1 << 1 // 对应 EXISTS
	Expire  = 1 << 2 // 对应 EXPIRE, EXPIREAT
	Persist = 1 << 3 // 对应 PERSIST
	TTL     = 1 << 4 // 对应 TTL, PTTL
	Type    = 1 << 5 // 对应 TYPE
	Rename  = 1 << 6 // 对应 RENAME, RENAMEX

	// 基础读写掩码 (无类型常量，自动适配所有 Op)
	CommonRead  = Exists | TTL | Type
	CommonWrite = Del | Expire | Persist | Rename
)

// -----------------------------------------------------------------------------
//  2. 各类型特定权限 (从 1<<10 开始复用位)
// -----------------------------------------------------------------------------

// Hash 权限
type HashOp uint64

const (
	HGet HashOp = 1 << (10 + iota)
	HSet
	HDel
	HMGET
	HExists
	HGetAll
	HRandField
	HRandFieldWithValues
	HLen
	HKeys
	HVals
	HIncrBy
	HIncrByFloat
	HSetNX
	HScan

	HashRead  = uint64(HGet|HMGET|HExists|HGetAll|HRandField|HRandFieldWithValues|HLen|HKeys|HVals|HScan) | CommonRead
	HashWrite = uint64(HSet|HDel|HIncrBy|HIncrByFloat|HSetNX) | CommonWrite
	HashAll   = HashRead | HashWrite
)

// List 权限
type ListOp uint64

const (
	RPush ListOp = 1 << (10 + iota)
	RPushX
	LPush
	LPushX
	RPop
	LPop
	LRange
	LRem
	LSet
	LIndex
	LTrim
	LLen

	ListRead  = uint64(LRange|LIndex|LLen) | CommonRead
	ListWrite = uint64(RPush|RPushX|LPush|LPushX|RPop|LPop|LRem|LSet|LTrim) | CommonWrite
	ListAll   = ListRead | ListWrite
)

// Set 权限
type SetOp uint64

const (
	SAdd SetOp = 1 << (10 + iota)
	SCard
	SRem
	SIsMember
	SMembers
	SScan

	SetRead  = uint64(SCard|SIsMember|SMembers|SScan) | CommonRead
	SetWrite = uint64(SAdd|SRem) | CommonWrite
	SetAll   = SetRead | SetWrite
)

// ZSet 权限 (已补全缺失的操作)
type ZSetOp uint64

const (
	ZAdd ZSetOp = 1 << (10 + iota)
	ZRem
	ZRange
	ZRank
	ZScore
	ZCard
	ZCount
	ZIncrBy
	ZScan
	ZRangeByScore
	ZRevRange
	ZRevRangeByScore
	ZRemRangeByScore
	ZRangeWithScores
	ZRevRangeWithScores

	// 更新 Read 掩码以包含补全的操作
	ZSetRead = uint64(ZRange|ZRank|ZScore|ZCard|ZCount|ZScan|
		ZRangeByScore|ZRevRange|ZRevRangeByScore|ZRangeWithScores|ZRevRangeWithScores) | CommonRead

	// 更新 Write 掩码
	ZSetWrite = uint64(ZAdd|ZRem|ZIncrBy|ZRemRangeByScore) | CommonWrite

	ZSetAll = ZSetRead | ZSetWrite
)

// String 权限
type StringOp uint64

const (
	Get StringOp = 1 << (10 + iota)
	Set
	StringGetAll
	StringSetAll

	StringRead  = uint64(Get|StringGetAll) | CommonRead
	StringWrite = uint64(Set|StringSetAll) | CommonWrite
	StringAll   = StringRead | StringWrite
)

// Stream 权限
type StreamOp uint64

const (
	XAdd StreamOp = 1 << (10 + iota)
	XDel
	XRange
	XLen
	XRead
	XTrim
	XInfo

	StreamRead  = uint64(XRange|XLen|XRead|XInfo) | CommonRead
	StreamWrite = uint64(XAdd|XDel|XTrim) | CommonWrite
	StreamAll   = StreamRead | StreamWrite
)

// VectorSet (FT.*) 权限
type VectorSetOp uint64

const (
	FtCreate VectorSetOp = 1 << (10 + iota)
	FtSearch
	FtAggregate
	FtDropIndex
	FtTagVals
	FtInfo

	VectorSetRead  = uint64(FtSearch|FtAggregate|FtTagVals|FtInfo) | CommonRead
	VectorSetWrite = uint64(FtCreate|FtDropIndex) | CommonWrite
	VectorSetAll   = VectorSetRead | VectorSetWrite
)

// -----------------------------------------------------------------------------
//  3. 系统级权限 (针对全局命令)
// -----------------------------------------------------------------------------

type DBOp uint64

const (
	DBTime DBOp = 1 << 10 // 对应 TIME
	DBKeys DBOp = 1 << 11 // 对应 KEYS
)

// -----------------------------------------------------------------------------
//  4. 权限逻辑实现
// -----------------------------------------------------------------------------

var HttpPermissions = cmap.New[uint64]()

// 底层校验：检查 Key 的掩码是否包含该操作位
func isHttpOpAllowed(key string, op uint64) bool {
	scope := strings.ToLower(KeyScope(key))
	mask, ok := HttpPermissions.Get(scope)
	return ok && (mask&op) != 0
}

// --- 暴露给 API 层的校验接口 ---

func IsAllowedHashOp(key string, op HashOp) bool           { return isHttpOpAllowed(key, uint64(op)) }
func IsAllowedListOp(key string, op ListOp) bool           { return isHttpOpAllowed(key, uint64(op)) }
func IsAllowedSetOp(key string, op SetOp) bool             { return isHttpOpAllowed(key, uint64(op)) }
func IsAllowedZSetOp(key string, op ZSetOp) bool           { return isHttpOpAllowed(key, uint64(op)) }
func IsAllowedStringOp(key string, op StringOp) bool       { return isHttpOpAllowed(key, uint64(op)) }
func IsAllowedStreamOp(key string, op StreamOp) bool       { return isHttpOpAllowed(key, uint64(op)) }
func IsAllowedVectorSetOp(key string, op VectorSetOp) bool { return isHttpOpAllowed(key, uint64(op)) }

// 通用生命周期校验 (如 DEL, EXPIRE 直接调用)
// op 传入无类型常量 (如 redisdb.Del)
func IsAllowedCommon(key string, op uint64) bool { return isHttpOpAllowed(key, op) }

// 全局 DB 校验 (强制检查 _systemdb 键)
func IsAllowedDBOp(op DBOp) bool { return isHttpOpAllowed(SystemDbKey, uint64(op)) }

// --- 权限设置接口 ---

func httpAllow(key string, op uint64) {
	scope := strings.ToLower(KeyScope(key))
	mask, exists := HttpPermissions.Get(scope)
	if exists {
		logger.Warn().Str("key", key).Msgf("overwriting existing HttpPermission mask 0x%X with 0x%X", mask, mask|op)
	}
	HttpPermissions.Set(scope, mask|op)
}

func AllowHashOp(key string, op uint64)      { httpAllow(key, op) }
func AllowListOp(key string, op uint64)      { httpAllow(key, op) }
func AllowSetOp(key string, op uint64)       { httpAllow(key, op) }
func AllowZSetOp(key string, op uint64)      { httpAllow(key, op) }
func AllowStringOp(key string, op uint64)    { httpAllow(key, op) }
func AllowStreamOp(key string, op uint64)    { httpAllow(key, op) }
func AllowVectorSetOp(key string, op uint64) { httpAllow(key, op) }

// AllowDBOp 设置全局系统权限
func AllowDBOp(op DBOp) { httpAllow(SystemDbKey, uint64(op)) }

// scope of a redis key (prefix before ':')
func KeyScope(key string) string {
	if before, _, found := strings.Cut(key, ":"); found {
		return before
	}
	return key
}
