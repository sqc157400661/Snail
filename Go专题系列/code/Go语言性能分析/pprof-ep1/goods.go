package main

import (
	"encoding/json"
	"fmt"
	jsoniter "github.com/json-iterator/go"
)

type Goods struct {
	ID         int     `json:"id"`
	GoodsName  string  `json:"goods_name"`
	GoodsPrice float64 `json:"goods_price"`
	GoodsColor string  `json:"goods_color"`
	Desc       string  `json:"Goods_desc"`
	StoreId    int     `json:"store_id"`
}

func getGoods1() Goods {

	//json字符中的"引号，需用\进行转义，否则编译出错
	data := "{\"id\":11,\"goods_name\":\"华为手机 荣耀40 pro\",\"goods_price\":6000.00,\"goods_color\":\"黑色\",\"Goods_desc\":\"测试测试测试测试测试测试\",\"StoreId\":65656}"
	str := []byte(data)

	//1.Unmarshal的第一个参数是json字符串，第二个参数是接受json解析的数据结构。
	//第二个参数必须是指针，否则无法接收解析的数据
	goods := Goods{}
	err := json.Unmarshal(str, &goods)

	//解析失败会报错，如json字符串格式不对，缺"号，缺}等。
	if err != nil {
		fmt.Println(err)
	}
	return goods
}

// 优化
func getGoods2() Goods {

	//json字符中的"引号，需用\进行转义，否则编译出错
	data := "{\"goods_name\":\"华为手机 荣耀40 pro\",\"goods_price\":6000.00,\"goods_color\":\"黑色\",\"Goods_desc\":\"测试测试测试测试测试测试\",\"StoreId\":65656}"
	str := []byte(data)

	//1.Unmarshal的第一个参数是json字符串，第二个参数是接受json解析的数据结构。
	//第二个参数必须是指针，否则无法接收解析的数据
	var newJson = jsoniter.ConfigCompatibleWithStandardLibrary
	goods := Goods{}
	err := newJson.Unmarshal(str, &goods)

	//解析失败会报错，如json字符串格式不对，缺"号，缺}等。
	if err != nil {
		fmt.Println(err)
	}
	return goods
}
