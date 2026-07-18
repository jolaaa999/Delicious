package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/delicious/delicious/pkg/cache"
)

// 常见中文菜名 → 英文检索词（公开菜谱 API 以英文为主）
var chineseDishMap = map[string]string{
	"排骨":   "pork ribs",
	"红烧肉":  "red braised pork belly",
	"番茄炒蛋": "tomato scrambled eggs",
	"番茄鸡蛋": "tomato scrambled eggs",
	"宫保鸡丁": "kung pao chicken",
	"鱼香肉丝": "yu xiang pork",
	"麻婆豆腐": "mapo tofu",
	"清蒸鲈鱼": "steamed sea bass",
	"糖醋里脊": "sweet and sour pork",
	"可乐鸡翅": "cola chicken wings",
	"回锅肉":  "twice cooked pork",
	"水煮鱼":  "boiled fish in chili oil",
	"酸菜鱼":  "pickled cabbage fish",
	"小炒肉":  "stir fried pork",
	"蛋炒饭":  "egg fried rice",
	"扬州炒饭": "yangzhou fried rice",
	"饺子":   "dumplings",
	"馄饨":   "wonton soup",
	"面条":   "noodles",
	"火锅":   "hot pot",
	"牛排":   "steak",
	"牛肉":   "beef",
}

func expandSearchKeywords(ctx context.Context, client *http.Client, keyword string) []string {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil
	}
	seen := map[string]bool{keyword: true}
	out := []string{keyword}

	if !containsChinese(keyword) {
		return out
	}

	if en, ok := chineseDishMap[keyword]; ok && en != "" && !seen[en] {
		seen[en] = true
		out = append(out, en)
	}

	if en, err := translateZhToEn(ctx, client, keyword); err == nil {
		en = strings.TrimSpace(en)
		if en != "" && !seen[en] {
			seen[en] = true
			out = append(out, en)
		}
	}

	return out
}

func translateZhToEn(ctx context.Context, client *http.Client, text string) (string, error) {
	return translateWithPair(ctx, client, text, "zh-CN|en")
}

func translateEnToZh(ctx context.Context, client *http.Client, text string) (string, error) {
	return translateWithPair(ctx, client, text, "en|zh-CN")
}

// ── 带缓存的翻译 ──

// CachedTranslate 带缓存的翻译，key 格式: "langPair:text"
func CachedTranslate(ctx context.Context, c *cache.MemoryCache, client *http.Client, text, langPair string) (string, error) {
	key := langPair + ":" + text
	if val, ok := c.Get(key); ok {
		return val, nil
	}
	val, err := translateWithPair(ctx, client, text, langPair)
	if err != nil {
		return "", err
	}
	c.Set(key, val)
	return val, nil
}

// CachedTranslateLong 带缓存的分段翻译
func CachedTranslateLong(ctx context.Context, c *cache.MemoryCache, client *http.Client, text, langPair string) (string, error) {
	key := "long:" + langPair + ":" + text
	if val, ok := c.Get(key); ok {
		return val, nil
	}
	val, err := translateLongText(ctx, client, text, langPair)
	if err != nil {
		return text, err
	}
	c.Set(key, val)
	return val, nil
}

func translateWithPair(ctx context.Context, client *http.Client, text, langPair string) (string, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return text, nil
	}
	if client == nil {
		client = http.DefaultClient
	}
	rawURL := fmt.Sprintf(
		"https://api.mymemory.translated.net/get?q=%s&langpair=%s",
		url.QueryEscape(text),
		langPair,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", fmt.Errorf("translate: HTTP %d", res.StatusCode)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	var parsed struct {
		ResponseData struct {
			TranslatedText string `json:"translatedText"`
		} `json:"responseData"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	translated := strings.TrimSpace(parsed.ResponseData.TranslatedText)
	if translated == "" || strings.EqualFold(translated, text) {
		return "", fmt.Errorf("empty translation")
	}
	return translated, nil
}

// translateLongText 分段翻译，避免免费 API 单条长度限制。
func translateLongText(ctx context.Context, client *http.Client, text, langPair string) (string, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return text, nil
	}
	const chunkSize = 400
	if len(text) <= chunkSize {
		out, err := translateWithPair(ctx, client, text, langPair)
		if err != nil {
			return text, err
		}
		return out, nil
	}
	parts := splitTextChunks(text, chunkSize)
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		translated, err := translateWithPair(ctx, client, part, langPair)
		if err != nil {
			out = append(out, part)
			continue
		}
		out = append(out, translated)
	}
	return strings.Join(out, "\n"), nil
}

func splitTextChunks(text string, maxLen int) []string {
	if len(text) <= maxLen {
		return []string{text}
	}
	var chunks []string
	runes := []rune(text)
	for i := 0; i < len(runes); i += maxLen {
		end := i + maxLen
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[i:end]))
	}
	return chunks
}
