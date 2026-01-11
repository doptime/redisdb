package redisdb

type IHttpKey interface {
	GetKeyType() KeyType
	GetUseModer() bool
	GetValue() interface{}
	ValidDataKey() error
	TimestampFiller(in interface{}) (err error)
}
