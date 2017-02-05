package radix

import (
	. "testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCmdAction(t *T) {
	c := dial()
	key, val := randStr(), randStr()

	require.Nil(t, Cmd(nil, "SET", key, val).Run(c))
	var got string
	require.Nil(t, Cmd(&got, "GET", key).Run(c))
	assert.Equal(t, val, got)
}

func TestLuaAction(t *T) {
	getset := `
		local res = redis.call("GET", KEYS[1])
		redis.call("SET", KEYS[1], ARGV[1])
		return res
	`
	getset += " -- " + randStr() // so it does has to do an eval every time

	c := dial()
	key := randStr()
	val1, val2 := randStr(), randStr()

	{
		var res string
		err := Lua(&res, getset, []string{key}, val1).Run(c)
		require.Nil(t, err, "%s", err)
		assert.Empty(t, res)
	}

	{
		var res string
		err := Lua(&res, getset, []string{key}, val2).Run(c)
		require.Nil(t, err)
		assert.Equal(t, val1, res)
	}
}

func TestPipelineAction(t *T) {
	c := dial()
	for i := 0; i < 10; i++ {
		ss := []string{
			randStr(),
			randStr(),
			randStr(),
		}
		out := make([]string, len(ss))
		var cmds []CmdAction
		for i := range ss {
			cmds = append(cmds, CmdNoKey(&out[i], "ECHO", ss[i]))
		}
		require.Nil(t, Pipeline(cmds...).Run(c))

		for i := range ss {
			assert.Equal(t, ss[i], out[i])
		}
	}
}

func TestWithConnAction(t *T) {
	c := dial()
	k, v := randStr(), 10

	err := WithConn([]byte(k), func(conn Conn) error {
		require.Nil(t, Cmd(nil, "SET", k, v).Run(conn))
		var out int
		require.Nil(t, Cmd(&out, "GET", k).Run(conn))
		assert.Equal(t, v, out)
		return nil
	}).Run(c)
	require.Nil(t, err)
}
