package client

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient_VUBFromHeight(t *testing.T) {
	t.Run("notary disabled", func(t *testing.T) {
		require.Zero(t, new(Client).VUBFromHeight(100))
	})

	c := &Client{notary: &notaryInfo{txValidTime: 50}}
	require.EqualValues(t, 199, c.VUBFromHeight(149))
	require.EqualValues(t, 201, c.VUBFromHeight(151))
}
