package service

// mapPriceAndDuration 根据检查类型映射价格和时长
func mapPriceAndDuration(category string) (int, string) {
	priceMap := map[string]int{
		"基因检测":  299,
		"血液检查":  50,
		"影像学检查": 200,
		"病理检查":  150,
		"生化检查":  80,
	}
	durationMap := map[string]string{
		"基因检测":  "3-5 个工作日",
		"血液检查":  "当天出结果",
		"影像学检查": "1-2 个工作日",
		"病理检查":  "5-7 个工作日",
		"生化检查":  "当天出结果",
	}
	price, ok := priceMap[category]
	if !ok {
		price = 100
	}
	duration, ok := durationMap[category]
	if !ok {
		duration = "1-3 个工作日"
	}
	return price, duration
}
