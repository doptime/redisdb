// package do stands for data options
package redisdb

// Option is parameter to create an API, RPC, or CallAt
type Option struct {
	RedisKey        string
	KeyType         KeyType
	RedisDataSource string
	HttpAccess      bool
	Modifiers       map[string]ModifierFunc
}

var Opt = Option{
	RedisKey:        "",
	RedisDataSource: "",
	HttpAccess:      false,
	Modifiers:       map[string]ModifierFunc{},
}

func (i Option) cp(o *Option) {
	o.RedisKey = i.RedisKey
	o.RedisDataSource = i.RedisDataSource
	o.HttpAccess = i.HttpAccess
	o.Modifiers = map[string]ModifierFunc{}
	for k, v := range i.Modifiers {
		o.Modifiers[k] = v
	}
}

func (i Option) Key(key string) (o Option) {
	i.cp(&o)
	o.RedisKey = key
	return
}

func (i Option) Rds(dataSource string) (o Option) {
	i.cp(&o)
	o.RedisDataSource = dataSource
	return
}

func (i Option) HttpVisit() (o Option) {
	i.cp(&o)
	o.HttpAccess = true
	return
}
func (i Option) Modifier(extraModifiers map[string]ModifierFunc) (o Option) {
	i.cp(&o)
	for k, v := range extraModifiers {
		o.Modifiers[k] = v
	}
	return
}
