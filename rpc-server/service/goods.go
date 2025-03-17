package service

type GoodsService struct {
	Find func(args map[string]any) ([]byte, error) `msrpc:"GET,/goods/find"`
}
