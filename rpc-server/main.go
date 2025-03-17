package main

import (
	"frame"
	"net/http"
	"rpc-client/model"
)

func main() {
	engine := frame.Default()
	g := engine.Group("goods")
	g.Get("/find", func(ctx *frame.Context) {
		//查询商品
		goods := model.Goods{Id: 1000, Name: "商品中心9001商品"}
		ctx.JSON(http.StatusOK, &model.Result{Code: 200, Msg: "success", Data: goods})
	})
	engine.Run(":9002")
}
