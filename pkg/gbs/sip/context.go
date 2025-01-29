package sip

type HandlerFunc func(*Context)

type Context struct {
	Request  *Request
	Tx       *Transaction
	handlers []HandlerFunc
	index    int

	cache map[string]any
}

func newContext(req *Request, tx *Transaction) *Context {
	return &Context{
		Request: req,
		Tx:      tx,
		cache:   make(map[string]any),
	}
}

func (c *Context) Next() {
	l := len(c.handlers)
	for {
		c.handlers[c.index](c)
		if c.index == -1 || c.index >= l {
			return
		}
		c.index++
	}
}

func (c *Context) Abort() {
	c.index = -1
}

func (c *Context) String(status int, msg string) {
	_ = c.Tx.Respond(NewResponseFromRequest("", c.Request, status, msg, nil))
}

func (c *Context) Set(k string, v any) {
	c.cache[k] = v
}

func (c *Context) Get(k string) (any, bool) {
	v, ok := c.cache[k]
	return v, ok
}

func (c *Context) GetMustString(k string) string {
	if v, ok := c.cache[k]; ok {
		return v.(string)
	}
	return ""
}

func (c *Context) GetMustInt(k string) int {
	if v, ok := c.cache[k]; ok {
		return v.(int)
	}
	return 0
}
