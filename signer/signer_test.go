package signer

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test1(t *testing.T) {
	signer := NewSinger("12345678", "helloworld")

	values := url.Values{}

	values.Set("session", "test")
	values.Set("method", "taobao.item.seller.get")
	values.Set("timestamp", "2016-01-01 12:00:00")
	values.Set("format", "json")

	values.Set("fields", "num_iid,title,nick,price,num")
	values.Set("num_iid", "11223344")

	values.Set("sign", "11")

	err := signer.Sign(values, SignMethodMD5)
	assert.Nil(t, err)
	assert.Equal(t, values.Get("sign"), "66987CB115214E59E6EC978214934FB8")
}
