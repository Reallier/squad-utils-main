package handler

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// 测试分割没有值的环境变量
func TestSpiltDockerEnvWithoutValue(t *testing.T) {
	env := "TEST="
	key, value, _, _ := SpiltDockerEnv(env)
	//fmt.Println(key)
	//fmt.Println(value)
	assert.Equal(t, key, "TEST")
	assert.Equal(t, len(value), 0)
}

// 测试分割没有 = 的环境变量
func TestSpiltDockerEnvWithValue(t *testing.T) {
	env := "TEST"
	key, value, length, err := SpiltDockerEnv(env)
	assert.Less(t, length, 2)
	assert.Equal(t, key, "")
	assert.Equal(t, value, "")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrEnvHasNoEqual)

}
