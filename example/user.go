package example

type User interface {
	// 获取用户
	GetUser(request Request) Response
}

type Request struct {
	Name string `json:"name"` // 用户名
}
type Response struct {
	Result string `json:"result"` // 返回结果
}
