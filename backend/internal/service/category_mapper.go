package service

import "strings"

// enToCnCategory 英文分类 → 中文分类映射。
// 外部 API 返回的原始分类统一转中文，确保分类筛选可用。
var enToCnCategory = map[string]string{
	"chinese":          "家常菜",
	"asian":            "家常菜",
	"japanese":         "日料",
	"korean":           "韩餐",
	"italian":          "西餐",
	"french":           "西餐",
	"american":         "西餐",
	"british":          "西餐",
	"mediterranean":    "西餐",
	"european":         "西餐",
	"mexican":          "西餐",
	"indian":           "东南亚",
	"thai":             "东南亚",
	"vietnamese":       "东南亚",
	"middle eastern":   "东南亚",
	"seafood":          "粤菜",
	"beef":             "硬菜",
	"chicken":          "家常菜",
	"pork":             "家常菜",
	"lamb":             "硬菜",
	"pasta":            "西餐",
	"vegetarian":       "素食",
	"vegan":            "素食",
	"dessert":          "甜点",
	"breakfast":        "早餐",
	"soup":             "汤羹",
	"salad":            "凉菜",
	"appetizer":        "凉菜",
	"snack":            "小吃",
	"beverage":         "饮品",
	"drink":            "饮品",
	"side dish":        "家常菜",
	"main course":      "硬菜",
	"fingerfood":       "小吃",
	"marinade":         "家常菜",
	"dip":              "凉菜",
	"bread":            "烘焙",
	"pastry":           "烘焙",
	"baking":           "烘焙",
	"stew":             "汤羹",
	"roast":            "硬菜",
	"grill":            "硬菜",
	"bbq":              "硬菜",
	"stir fry":         "家常菜",
	"steam":            "家常菜",
	"miscellaneous":    "家常菜",
	"unknown":          "家常菜",
}

// enToCnTag 英文标签 → 中文标签映射
var enToCnTag = map[string]string{
	"chinese":      "家常菜",
	"spicy":        "麻辣",
	"soup":         "汤羹",
	"vegetarian":   "素食",
	"vegan":        "素食",
	"seafood":      "粤菜",
	"breakfast":    "早餐",
	"dessert":      "甜点",
	"beef":         "硬菜",
	"chicken":      "家常菜",
	"pork":         "家常菜",
	"pasta":        "西餐",
	"salad":        "凉菜",
	"appetizer":    "凉菜",
	"snack":        "小吃",
	"quick":        "快手菜",
	"easy":         "快手菜",
	"healthy":      "低脂",
	"low fat":      "低脂",
	"high protein": "高蛋白",
	"comfort food": "家常菜",
	"traditional":  "家常菜",
	"gourmet":      "宴客菜",
	"dinner":       "晚餐",
	"lunch":        "午餐",
	"side":         "家常菜",
	"main dish":    "硬菜",
}

// MapCategory 将英文分类转为中文。无匹配时保留原值。
func MapCategory(en string) string {
	key := strings.ToLower(strings.TrimSpace(en))
	if cn, ok := enToCnCategory[key]; ok {
		return cn
	}
	return en
}

// MapTag 将英文标签转为中文标签。无匹配时保留原值。
func MapTag(en string) string {
	key := strings.ToLower(strings.TrimSpace(en))
	if cn, ok := enToCnTag[key]; ok {
		return cn
	}
	if cn, ok := enToCnCategory[key]; ok {
		return cn
	}
	return en
}

// MapTags 批量映射标签，去重去空。
func MapTags(tags []string) []string {
	seen := map[string]bool{}
	var result []string
	for _, t := range tags {
		mapped := MapTag(t)
		if mapped == "" || seen[mapped] {
			continue
		}
		seen[mapped] = true
		result = append(result, mapped)
	}
	return result
}
