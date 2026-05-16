// ============================================================
// LyricCanvas — 押韵词联想服务
// ============================================================
package services

import (
	_ "embed"
	"encoding/json"
	"strings"
	"sync"

	"lyriccanvas/models"
)

//go:embed pinyin_dict.json
var pinyinDictRaw []byte

type RhymeService struct {
	mu         sync.RWMutex
	charPinyin map[rune]string   // 汉字 → 拼音
	finalChars map[string][]rune // 韵母 → 汉字列表
}

func NewRhymeService() (*RhymeService, error) {
	s := &RhymeService{
		charPinyin: make(map[rune]string),
		finalChars: make(map[string][]rune),
	}

	var raw map[string]string
	if err := json.Unmarshal(pinyinDictRaw, &raw); err != nil {
		return nil, err
	}

	for charStr, pinyin := range raw {
		runes := []rune(charStr)
		if len(runes) != 1 {
			continue
		}
		r := runes[0]
		s.charPinyin[r] = pinyin
		final := extractFinal(pinyin)
		s.finalChars[final] = append(s.finalChars[final], r)
	}

	return s, nil
}

// extractFinal 从拼音提取韵母
func extractFinal(pinyin string) string {
	pinyin = strings.ToLower(pinyin)
	// 去掉声母
	final := pinyin
	for _, initial := range []string{"zh", "ch", "sh", "b", "p", "m", "f", "d", "t", "n", "l", "g", "k", "h", "j", "q", "x", "r", "z", "c", "s", "y", "w"} {
		if strings.HasPrefix(final, initial) {
			final = final[len(initial):]
			break
		}
	}
	// 保留韵母（含介音）
	return final
}

// Query 查询单字押韵词
func (s *RhymeService) Query(char string) (*models.RhymeResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	runes := []rune(char)
	if len(runes) == 0 {
		return &models.RhymeResult{Char: char}, nil
	}

	target := runes[len(runes)-1] // 取最后一个字
	pinyin, ok := s.charPinyin[target]
	if !ok {
		return &models.RhymeResult{Char: char, Pinyin: "?", Final: "?", Rhymes: []models.RhymeChar{}}, nil
	}

	final := extractFinal(pinyin)
	sameFinals := s.finalChars[final]

	var rhymes []models.RhymeChar
	seen := make(map[rune]bool)
	for _, r := range sameFinals {
		if r == target || seen[r] {
			continue
		}
		seen[r] = true
		rhymes = append(rhymes, models.RhymeChar{
			Char:   string(r),
			Pinyin: s.charPinyin[r],
		})
		// 限制返回数量
		if len(rhymes) >= 100 {
			break
		}
	}

	result := &models.RhymeResult{
		Char:   string(target),
		Pinyin: pinyin,
		Final:  final,
		Rhymes: rhymes,
	}

	return result, nil
}

// QueryBatch 批量查询
func (s *RhymeService) QueryBatch(chars []string) ([]models.RhymeResult, error) {
	results := make([]models.RhymeResult, len(chars))
	for i, ch := range chars {
		r, err := s.Query(ch)
		if err != nil {
			return nil, err
		}
		results[i] = *r
	}
	return results, nil
}
