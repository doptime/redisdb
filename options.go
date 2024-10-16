// package do stands for data options
package redisdb

// Option is parameter to create an API, RPC, or CallAt
type Option struct {
	Key        string
	DataSource string
	AsWebData  bool
	Modifiers  map[string]ModifierFunc
}
type opSetter func(*Option)

func WithKey(key string) opSetter {
	return func(o *Option) {
		o.Key = key
	}
}

func WithRds(dataSource string) opSetter {
	return func(o *Option) {
		o.DataSource = dataSource
	}
}

func WithAsWebData() opSetter {
	return func(o *Option) {
		o.AsWebData = true
	}
}
func WithModifier(extraModifiers map[string]ModifierFunc) opSetter {
	return func(o *Option) {
		o.Modifiers = extraModifiers
	}
}

func (opt Option) buildOptions(options ...opSetter) *Option {
	for _, option := range options {
		option(&opt)
	}
	return &opt
}
