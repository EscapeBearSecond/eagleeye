package eagleeye

// Option 选项
type Option interface {
	apply(*EagleeyeEngine)
}

// fn 函数类型
type fn func(*EagleeyeEngine)

// apply 实现Option接口
func (f fn) apply(e *EagleeyeEngine) {
	f(e)
}

// WithDirectory 配置输出目录
func WithDirectory(dir string) Option {
	return fn(func(e *EagleeyeEngine) {
		e.dir = dir
	})
}

type extraOptions struct {
	id string
}

type ExtraOption interface {
	apply(*extraOptions)
}

type extraFn func(*extraOptions)

func (f extraFn) apply(e *extraOptions) {
	f(e)
}

func WithID(id string) ExtraOption {
	return extraFn(func(e *extraOptions) {
		e.id = id
	})
}
