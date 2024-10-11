package tpl

// options 选项
type options struct {
	headless bool
}

// Option 选项函数
type Option func(*options)

// WithEnableHeadless 配置是否启用headless模式
func WithEnableHeadless(headless bool) Option {
	return func(o *options) {
		o.headless = headless
	}
}
