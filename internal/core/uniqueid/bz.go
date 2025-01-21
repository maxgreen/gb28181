// uniqueid
// 的设计是用于生成全局唯一的 ID，避免重复。
// 此库不考虑分布式，仅通过数据库主键索引来实现。
// 当 id 重复时，由业务端抛出错误即可

package uniqueid

import (
	"context"
	"crypto/rand"
	"log/slog"
	"math/big"
)

type IDManager struct {
	store  UniqueIDStorer
	length int
	// TODO: 可以初始化时读取数据库内的数量，判断重复因子，从而减少尝试或更换策略
}

func NewIDManager(store UniqueIDStorer, length int) *IDManager {
	return &IDManager{
		store:  store,
		length: length,
	}
}

func (m *IDManager) UniqueID() string {
	for i := range 10 {
		for range 66 {
			id := GenerateRandomString(m.length + i)
			if err := m.store.Add(context.Background(), &UniqueID{ID: id}); err != nil {
				slog.Error("UniqueID", "err", err)
				continue
			}
			return id
		}
	}
	slog.Error("UniqueID", "err", "超过最大循环次数，未获取到唯一 id")
	return "unknown"
}

// GenerateRandomString 生成随机字符串
func GenerateRandomString(length int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyz1234567890"
	lettersLength := big.NewInt(int64(len(letterBytes)))
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		idx, _ := rand.Int(rand.Reader, lettersLength)
		result[i] = letterBytes[idx.Int64()]
	}
	return string(result)
}
