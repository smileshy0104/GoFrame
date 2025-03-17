package main

import (
	"encoding/json"
	"fmt"
	"frame"
	"frame/rpc"
	"net/http"
	"rpc-client/model"
)

func main() {
	engine := frame.Default()
	client := rpc.NewHttpClient()
	g := engine.Group("order")
	g.Get("/find", func(ctx *frame.Context) {
		//查询商品
		bytes, err := client.Session().Get("http://localhost:9002/goods/find", nil)
		//bytes, err := client.Get("http://localhost:9002/goods/find", nil)
		if err != nil {
			ctx.Logger.Error(err)
		}
		fmt.Println(string(bytes))
		v := &model.Result{}
		json.Unmarshal(bytes, v)
		ctx.JSON(http.StatusOK, v)
	})
	engine.Run(":9003")
}
